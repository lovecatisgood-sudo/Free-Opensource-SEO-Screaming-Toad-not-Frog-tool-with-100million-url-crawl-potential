export interface CrawlProgress { crawl_id: string; status: string; discovered: number; queued: number; fetched: number; analysed: number; failed: number; terminal_reason?: string; updated_at: string }
export interface CrawlResult { project_id: string; crawl_id: string; progress: CrawlProgress }
export interface Summary { crawl_id: string; status: string; discovered: number; fetched: number; analysed: number; failed: number; issues_by_severity: Record<string, number>; responses_by_class: Record<string, number>; rendering_by_status: Record<string, number> }
export interface PageRecord { id: number; url: string; title: string; meta_description: string; status_code: number; depth: number; canonical_url: string; extraction_mode:string; render_status?:string; rendered_title?:string; rendered_meta_description?:string; rendered_canonical_url?:string; rendered_content_hash?:string }
export interface IssueRecord { id: number; rule_id: string; rule_version: number; severity: string; evidence_json: string; subject_type: string; subject_id: string }
export interface Page<T> { items: T[]; next_cursor?: string }
export interface Project { project_id: string; name: string; archived: boolean; created_at: string; updated_at: string }
export interface CrawlLimits { maximum_urls: number; maximum_depth: number; maximum_duration: number; maximum_body_bytes: number; maximum_disk_bytes: number; global_concurrency: number; per_host_concurrency: number; minimum_host_delay: number }
export interface CrawlConfiguration { seed_url: string; allowed_hosts: string[]; allow_subdomains: boolean; include_path_regex?: string[]; exclude_path_regex?: string[]; include_query_regex?: string[]; exclude_query_regex?: string[]; user_agent?: string; rendering_mode: "raw" | "rendered"; limits: CrawlLimits }
export interface Profile { profile_id: string; project_id: string; version: number; name: string; configuration: CrawlConfiguration; created_at: string }
export interface ScopeDecision { url: string; normalized_url?: string; allowed: boolean; reason?: string }
export interface Comparison { base_crawl_id: string; target_crawl_id: string; configuration_match: boolean; added_pages: number; removed_pages: number; changed_pages: number; new_issues: number; fixed_issues: number }
export interface RenderedPage { status:string; error_code?:string; final_url?:string; request_count:number; transferred_bytes:number; title?:string; meta_description?:string; canonical_url?:string; text_length:number; headings:Array<{level:number;text:string}> }
export interface RenderDifference { field:string; raw_value:string; rendered_value:string }
export interface PageDetail { page: PageRecord; headings: Array<{level:number;text:string}>; inlinks: unknown[]; outlinks: unknown[]; images: unknown[]; hreflang: unknown[]; structured_data: unknown[]; issues: IssueRecord[]; rendered?:RenderedPage; render_differences:RenderDifference[] }
export interface IssueExplanation { issue: IssueRecord; rule: { title:string; category:string; remediation:string; limitations:string; version:number; default_severity:string } }

let csrfToken = "";
async function request<T>(path: string, init?: RequestInit): Promise<T> { const response = await fetch(path, { credentials: "same-origin", ...init, headers: { "Accept": "application/json", ...(init?.body ? { "Content-Type": "application/json" } : {}), ...(init?.method && init.method !== "GET" ? { "X-CSRF-Token": csrfToken } : {}), ...init?.headers } }); if (!response.ok) { const body = await response.json().catch(() => null) as { error?: { message?: string } } | null; throw new Error(body?.error?.message ?? `Request failed (${response.status})`); } if (response.status === 204) return undefined as T; return response.json() as Promise<T>; }
export async function bootstrap(): Promise<void> { const value = await request<{ csrf_token: string }>("/api/v1/session", { method: "POST" }); csrfToken = value.csrf_token; }
export function createProject(name:string){return request<Project>("/api/v1/projects",{method:"POST",body:JSON.stringify({name})});}
export function listProjects(){return request<Page<Project>>("/api/v1/projects?limit=100");}
export function createProfile(projectID:string,name:string,configuration:CrawlConfiguration){return request<Profile>(`/api/v1/projects/${encodeURIComponent(projectID)}/profiles`,{method:"POST",body:JSON.stringify({name,configuration})});}
export function listProfiles(projectID:string){return request<Page<Profile>>(`/api/v1/projects/${encodeURIComponent(projectID)}/profiles?limit=100`);}
export function previewScope(projectID:string,configuration:CrawlConfiguration,urls:string[]){return request<{decisions:ScopeDecision[]}>(`/api/v1/projects/${encodeURIComponent(projectID)}/scope-preview`,{method:"POST",body:JSON.stringify({configuration,urls})});}
export function startProfileCrawl(projectID:string,profileID:string){return request<CrawlResult>(`/api/v1/projects/${encodeURIComponent(projectID)}/crawls`,{method:"POST",body:JSON.stringify({profile_id:profileID})});}
export function listCrawls(projectID:string){return request<Page<CrawlProgress>>(`/api/v1/projects/${encodeURIComponent(projectID)}/crawls?limit=100`);}
export function startCrawl(input: { url: string; name: string; allow_subdomains: boolean; maximum_urls: number }): Promise<CrawlResult> { return request("/api/v1/crawls", { method: "POST", body: JSON.stringify(input) }); }
export function crawlStatus(id: string): Promise<CrawlProgress> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/status`); }
export function auditSummary(id: string): Promise<Summary> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/summary`); }
export function pages(id: string, search = "", cursor=""): Promise<Page<PageRecord>> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/pages?limit=100&search=${encodeURIComponent(search)}&cursor=${encodeURIComponent(cursor)}`); }
export function issues(id: string, search = "", cursor=""): Promise<Page<IssueRecord>> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/issues?limit=100&search=${encodeURIComponent(search)}&cursor=${encodeURIComponent(cursor)}`); }
export function pageDetail(crawlID:string,pageID:number){return request<PageDetail>(`/api/v1/crawls/${encodeURIComponent(crawlID)}/pages/${pageID}`);}
export function explainIssue(crawlID:string,issueID:number){return request<IssueExplanation>(`/api/v1/crawls/${encodeURIComponent(crawlID)}/issues/${issueID}`);}
export function cancelCrawl(id: string): Promise<void> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/cancel`, { method: "POST" }); }
export function pauseCrawl(id: string): Promise<void> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/pause`, { method: "POST" }); }
export function resumeCrawl(id: string): Promise<void> { return request(`/api/v1/crawls/${encodeURIComponent(id)}/resume`, { method: "POST" }); }
export function compareCrawls(base:string,target:string){return request<Comparison>("/api/v1/comparisons",{method:"POST",body:JSON.stringify({base_crawl_id:base,target_crawl_id:target})});}
export function createExport(crawlID: string, dataset: "pages" | "issues" | "workbook", format: "csv" | "ndjson" | "xlsx") { return request<{ artifact_id: string; path: string }>("/api/v1/exports", { method: "POST", body: JSON.stringify({ crawl_id: crawlID, dataset, format }) }); }
