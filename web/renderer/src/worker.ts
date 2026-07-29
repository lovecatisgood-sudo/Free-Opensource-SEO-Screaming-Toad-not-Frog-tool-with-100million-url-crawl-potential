import { randomUUID } from "node:crypto";
import { once } from "node:events";
import { chromium, type Browser, type BrowserContext, type Route } from "playwright";
import {
  FrameDecoder,
  encodeFrame,
  maximumRenderedHTMLBytes,
  maximumResourceBytes,
  validateRenderRequest,
  type FetchResourceResponse,
  type RenderRequest,
  type RenderResponse,
  type SupervisorMessage,
  type WorkerMessage,
} from "./protocol.js";

const decoder = new FrameDecoder();
const pending = new Map<string, {
  resolve: (value: FetchResourceResponse) => void;
  reject: (reason: Error) => void;
}>();
let running = false;
let browser: Browser | null = null;
let pagesSinceRecycle = 0;

async function send(value: WorkerMessage): Promise<void> {
  if (!process.stdout.write(encodeFrame(value))) {
    await once(process.stdout, "drain");
  }
}

async function getBrowser(): Promise<Browser> {
  if (browser?.isConnected() && pagesSinceRecycle < 25) return browser;
  if (browser) await browser.close().catch(() => undefined);
  const containerSandbox = process.env.SEO_AUDITOR_CONTAINER_SANDBOX === "1";
  browser = await chromium.launch({
    headless: true,
    chromiumSandbox: !containerSandbox,
    args: [
      "--disable-background-networking",
      "--disable-component-update",
      "--disable-default-apps",
      "--disable-domain-reliability",
      "--disable-features=WebRtcHideLocalIpsWithMdns",
      "--disable-sync",
      "--disable-webrtc",
      "--host-resolver-rules=MAP * ~NOTFOUND",
      "--metrics-recording-only",
      "--no-first-run",
      "--proxy-bypass-list=<-loopback>",
      "--proxy-server=http://127.0.0.1:9",
    ],
  });
  pagesSinceRecycle = 0;
  return browser;
}

function fetchThroughSupervisor(request: RenderRequest, route: Route, signal: AbortSignal): Promise<FetchResourceResponse> {
  const fetchId = randomUUID();
  return new Promise((resolve, reject) => {
    const abort = () => {
      pending.delete(fetchId);
      reject(new Error("render_deadline"));
    };
    const fail = (reason: unknown) => {
      pending.delete(fetchId);
      signal.removeEventListener("abort", abort);
      reject(reason instanceof Error ? reason : new Error("supervisor_write_failed"));
    };
    signal.addEventListener("abort", abort, { once: true });
    pending.set(fetchId, {
      resolve: (value) => {
        signal.removeEventListener("abort", abort);
        resolve(value);
      },
      reject: fail,
    });
    send({
      kind: "fetch_resource", protocolVersion: 1, requestId: request.requestId,
      fetchId, url: route.request().url(), resourceType: route.request().resourceType(),
    }).catch(fail);
  });
}

async function render(request: RenderRequest): Promise<RenderResponse> {
  let requestCount = 0;
  let transferredBytes = 0;
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), request.deadlineMs);
  let context: BrowserContext | null = null;
  try {
    const activeBrowser = await getBrowser();
    context = await activeBrowser.newContext({
      acceptDownloads: false,
      serviceWorkers: "block",
      javaScriptEnabled: true,
      bypassCSP: false,
      ignoreHTTPSErrors: false,
    });
    context.setDefaultNavigationTimeout(request.deadlineMs);
    context.setDefaultTimeout(Math.min(request.deadlineMs, 30_000));
    await context.addInitScript(`
      if (navigator.serviceWorker) {
        Object.defineProperty(navigator.serviceWorker, "register", {
          value: () => Promise.reject(new DOMException("Service workers are disabled", "NotAllowedError"))
        });
      }
    `);
    await context.routeWebSocket("**/*", (socket) => socket.close());
    await context.route("**/*", async (route) => {
      try {
        if (controller.signal.aborted) return route.abort("timedout");
        const resourceType = route.request().resourceType();
        if (resourceType === "media" || resourceType === "font") return route.abort("blockedbyclient");
        if (requestCount >= request.maximumRequests) return route.abort("blockedbyclient");
        requestCount++;
        const response = await fetchThroughSupervisor(request, route, controller.signal);
        if (response.status !== "completed") return route.abort("blockedbyclient");
        const body = Buffer.from(response.bodyBase64 ?? "", "base64");
        if (body.byteLength > maximumResourceBytes || transferredBytes + body.byteLength > request.maximumBytes) {
          return route.abort("blockedbyclient");
        }
        transferredBytes += body.byteLength;
        const headers = { ...(response.headers ?? {}) };
        delete headers["content-encoding"];
        delete headers["content-length"];
        delete headers["set-cookie"];
        await route.fulfill({ status: response.statusCode ?? 200, headers, body });
      } catch {
        return route.abort("failed");
      }
    });
    const page = await context.newPage();
    page.on("dialog", (dialog) => dialog.dismiss().catch(() => undefined));
    page.on("download", (download) => download.cancel().catch(() => undefined));
    await page.goto(request.url, { waitUntil: "domcontentloaded", timeout: request.deadlineMs });
    await page.waitForLoadState("networkidle", { timeout: Math.min(2_000, request.deadlineMs) }).catch(() => undefined);
    const html = await page.content();
    if (Buffer.byteLength(html, "utf8") > maximumRenderedHTMLBytes) throw new Error("rendered_html_limit");
    pagesSinceRecycle++;
    return {
      kind: "render_result", protocolVersion: 1, requestId: request.requestId,
      status: "completed", html, finalURL: page.url(), requestCount, transferredBytes,
    };
  } catch (reason) {
    return {
      kind: "render_result", protocolVersion: 1, requestId: request.requestId,
      status: controller.signal.aborted ? "failed" : "blocked", requestCount, transferredBytes,
      errorCode: reason instanceof Error ? reason.message : "render_failed",
    };
  } finally {
    clearTimeout(timeout);
    await context?.close().catch(() => undefined);
  }
}

async function receive(value: unknown): Promise<void> {
  const message = value as SupervisorMessage;
  if (message?.kind === "fetch_resource_result") {
    const waiter = pending.get(message.fetchId);
    if (waiter) {
      pending.delete(message.fetchId);
      waiter.resolve(message);
    }
    return;
  }
  let request: RenderRequest;
  try {
    request = validateRenderRequest(value);
  } catch (reason) {
    await send({
      kind: "render_result", protocolVersion: 1, requestId: "invalid",
      status: "failed", requestCount: 0, transferredBytes: 0,
      errorCode: reason instanceof Error ? reason.message : "invalid_request",
    });
    return;
  }
  if (running) {
    await send({
      kind: "render_result", protocolVersion: 1, requestId: request.requestId,
      status: "failed", requestCount: 0, transferredBytes: 0, errorCode: "worker_busy",
    });
    return;
  }
  running = true;
  try {
    await send(await render(request));
  } finally {
    running = false;
  }
}

process.stdin.on("data", (chunk) => {
  try {
    for (const value of decoder.push(chunk)) void receive(value);
  } catch (reason) {
    void send({
      kind: "render_result", protocolVersion: 1, requestId: "invalid",
      status: "failed", requestCount: 0, transferredBytes: 0,
      errorCode: reason instanceof Error ? reason.message : "protocol_error",
    });
  }
});
process.stdin.on("end", () => {
  void browser?.close().finally(() => process.exit(0));
});
