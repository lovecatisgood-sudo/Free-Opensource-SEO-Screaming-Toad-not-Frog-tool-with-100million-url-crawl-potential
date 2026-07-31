import { mkdir } from "node:fs/promises";
import { resolve } from "node:path";
import { chromium } from "../web/renderer/node_modules/playwright/index.mjs";

const baseURL = process.argv[2] ?? "http://127.0.0.1:7342";
const outputDirectory = resolve(process.argv[3] ?? ".artifacts/course-bonus-screenshots");
const projectName = process.argv[4] ?? "DJAI Academy course reality audit 2026-07-30";
const crawlID = process.argv[5] ?? "crawl_425921818cae33f4a03a417f7880cd3e";

await mkdir(outputDirectory, { recursive: true });

async function captureTop(page, locator, fileName, maximumHeight = 900) {
  const box = await locator.boundingBox();
  if (!box) throw new Error(`Could not resolve screenshot bounds for ${fileName}`);
  await page.screenshot({
    path: resolve(outputDirectory, fileName),
    clip: { x: box.x, y: box.y, width: box.width, height: Math.min(box.height, maximumHeight) },
  });
}

const browser = await chromium.launch({
  headless: true,
  executablePath: "/usr/bin/google-chrome",
  args: ["--disable-dev-shm-usage"],
});

try {
  const page = await browser.newPage({ viewport: { width: 1680, height: 1050 }, deviceScaleFactor: 1 });
  await page.goto(baseURL, { waitUntil: "networkidle" });
  await page.getByText("Local API ready", { exact: true }).waitFor();
  await page.getByRole("button", { name: new RegExp(projectName) }).click();
  await page.getByText("DJAI Academy raw reality audit - 100k ceiling", { exact: true }).waitFor();

  await page.getByRole("button", { name: new RegExp(crawlID) }).click();
  await page.getByRole("heading", { name: "Audit results" }).waitFor();

  await page.locator(".progress-panel").screenshot({ path: resolve(outputDirectory, "01-live-crawl-completed.png") });
  await captureTop(page, page.locator("#audit-results"), "02-live-audit-summary.png");

  const search = page.locator("#audit-results input[placeholder^='Filter URL']");
  await search.fill("noindex");
  await page.locator("#audit-results form").getByRole("button", { name: "Search" }).click();
  await page.locator("#audit-results tbody tr").first().waitFor();
  await captureTop(page, page.locator("#audit-results"), "03-noindex-review.png");

  await page.locator("#audit-results tbody tr").first().getByRole("button", { name: "Explain" }).click();
  await page.locator(".drawer").waitFor();
  await page.locator(".drawer").screenshot({ path: resolve(outputDirectory, "04-indexability-rule-explanation.png") });
  await page.getByRole("button", { name: "Close detail" }).click();

  await search.fill("hamming_distance");
  await page.locator("#audit-results form").getByRole("button", { name: "Search" }).click();
  await page.locator("#audit-results tbody tr").first().waitFor();
  await captureTop(page, page.locator("#audit-results"), "05-near-duplicate-review.png");

  await search.fill("");
  await page.locator("#audit-results form").getByRole("button", { name: "Search" }).click();
  await page.getByRole("tab", { name: /^Pages/ }).click();
  await page.locator("#audit-results tbody tr").first().waitFor();
  await captureTop(page, page.locator("#audit-results"), "06-live-page-inventory.png");

  await page.locator("#crawl-history").screenshot({ path: resolve(outputDirectory, "07-live-crawl-history.png") });
} finally {
  await browser.close();
}
