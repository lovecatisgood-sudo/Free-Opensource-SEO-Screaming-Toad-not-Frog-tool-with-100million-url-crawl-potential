export interface CrawlProgress { crawl_id: string; status: string; discovered: number; queued: number; fetched: number; analysed: number; failed: number; terminal_reason?: string }
export interface CrawlResult { project_id: string; crawl_id: string; progress: CrawlProgress }
export interface Summary { crawl_id: string; status: string; Discovered: number; Fetched: number; Analysed: number; Failed: number; issues_by_severity: Record<string, number>; responses_by_class: Record<string, number> }
export interface PageRecord { id: number; URL: string; Title: string; StatusCode: number; Depth: number; CanonicalURL: string }
export interface IssueRecord { id: number; rule_id: string; severity: string; EvidenceJSON: string }
export interface Page<T> { items: T[]; next_cursor?: string }

let csrfToken = "";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { credentials: "same-origin", ...init, headers: { "Accept": "application/json", ...(init?.body ? { "Content-Type": "application/json" } : {}), ...(init?.method && init.method !== "GET" ? { "X-CSRF-Token": csrfToken } : {}), ...init?.headers } });
  if (!response.ok) { const body = await response.json().catch(() => null) as { error?: { message?: string } } | null; throw new Error(body?.error?.message ?? `Request failed (${response.status})`); }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export async function bootstrap(): Promise<void> { const value = await request<{ csrf_token: string }>("/api/v1/session", { method: "POST" }); csrfToken = value.csrf_token; }
export function startCrawl(input: { url: string; name: string; allow_subdomains: boolean; maximum_urls: number }): Promise<CrawlResult> { return request("/api/v1/crawls", { method: "POST", body: JSON.stringify(input) }); }
export function crawlStatus(id: string): Promise<CrawlProgress> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/status`); }
export function auditSummary(id: string): Promise<Summary> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/summary`); }
export function pages(id: string): Promise<Page<PageRecord>> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/pages?limit=100`); }
export function issues(id: string): Promise<Page<IssueRecord>> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/issues?limit=100`); }
export function cancelCrawl(id: string): Promise<void> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/cancel`, { method: "POST" }); }

