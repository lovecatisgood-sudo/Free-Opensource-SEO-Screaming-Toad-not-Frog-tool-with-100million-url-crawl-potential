import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { auditSummary, bootstrap, cancelCrawl, crawlStatus, issues, pages, startCrawl, type CrawlProgress, type IssueRecord, type PageRecord, type Summary } from "./api";

const terminal = new Set(["completed", "cancelled", "failed", "limit_reached"]);

export function App() {
  const [ready, setReady] = useState(false); const [error, setError] = useState(""); const [busy, setBusy] = useState(false);
  const [url, setURL] = useState(""); const [name, setName] = useState("Website audit"); const [maximumURLs, setMaximumURLs] = useState(10_000); const [subdomains, setSubdomains] = useState(false);
  const [crawl, setCrawl] = useState<CrawlProgress | null>(null); const [summary, setSummary] = useState<Summary | null>(null); const [pageRows, setPageRows] = useState<PageRecord[]>([]); const [issueRows, setIssueRows] = useState<IssueRecord[]>([]); const [view, setView] = useState<"issues" | "pages">("issues");

  useEffect(() => { bootstrap().then(() => setReady(true)).catch((reason: unknown) => setError(reason instanceof Error ? reason.message : "Could not connect to the local API")); }, []);
  useEffect(() => {
    if (!crawl || terminal.has(crawl.status)) return;
    const timer = window.setInterval(() => { crawlStatus(crawl.crawl_id).then(setCrawl).catch((reason: unknown) => setError(reason instanceof Error ? reason.message : "Status request failed")); }, 1000);
    return () => window.clearInterval(timer);
  }, [crawl?.crawl_id, crawl?.status]);
  useEffect(() => {
    if (!crawl || !terminal.has(crawl.status)) return;
    Promise.all([auditSummary(crawl.crawl_id), pages(crawl.crawl_id), issues(crawl.crawl_id)]).then(([nextSummary, nextPages, nextIssues]) => { setSummary(nextSummary); setPageRows(nextPages.items); setIssueRows(nextIssues.items); }).catch((reason: unknown) => setError(reason instanceof Error ? reason.message : "Could not load audit results"));
  }, [crawl?.crawl_id, crawl?.status]);

  async function submit(event: FormEvent) { event.preventDefault(); setBusy(true); setError(""); setSummary(null); setPageRows([]); setIssueRows([]); try { const result = await startCrawl({ url, name, allow_subdomains: subdomains, maximum_urls: maximumURLs }); setCrawl(result.progress); } catch (reason) { setError(reason instanceof Error ? reason.message : "Could not start crawl"); } finally { setBusy(false); } }
  async function cancel() { if (!crawl) return; try { await cancelCrawl(crawl.crawl_id); setCrawl(await crawlStatus(crawl.crawl_id)); } catch (reason) { setError(reason instanceof Error ? reason.message : "Could not cancel crawl"); } }

  const progressMaximum = Math.max(crawl?.discovered ?? 1, 1); const progressValue = crawl?.analysed ?? 0;
  return <main className="shell">
    <header className="masthead"><div><p className="eyebrow">Local-first technical SEO</p><h1>SEO Auditor</h1><p className="lede">Crawl safely. Inspect the evidence. Fix what matters.</p></div><span className={`connection ${ready ? "online" : ""}`}>{ready ? "Local API ready" : "Connecting…"}</span></header>
    <section className="workspace" aria-labelledby="new-audit-heading">
      <div className="section-heading"><p className="kicker">New campaign</p><h2 id="new-audit-heading">Audit a public website</h2><p>Robots rules, redirects, DNS targets, response sizes, concurrency, and crawl limits are enforced automatically.</p></div>
      <form onSubmit={submit} className="crawl-form"><label className="wide">Website URL<input type="url" required placeholder="https://example.com/" value={url} onChange={(event) => setURL(event.target.value)} /></label><label>Campaign name<input required maxLength={200} value={name} onChange={(event) => setName(event.target.value)} /></label><label>URL ceiling<input type="number" min={1} max={100_000_000} value={maximumURLs} onChange={(event) => setMaximumURLs(Number(event.target.value))} /></label><label className="check"><input type="checkbox" checked={subdomains} onChange={(event) => setSubdomains(event.target.checked)} /><span>Include subdomains</span></label><button className="primary" disabled={!ready || busy}>{busy ? "Validating target…" : "Start audit"}</button></form>
      {error && <div className="alert" role="alert">{error}</div>}
    </section>
    {crawl && <section className="progress-panel" aria-labelledby="crawl-progress-heading"><div><p className="kicker">Campaign status</p><h2 id="crawl-progress-heading">{crawl.status.replaceAll("_", " ")}</h2><p className="mono">{crawl.crawl_id}</p></div><div className="progress-content"><progress max={progressMaximum} value={progressValue}>{progressValue} of {progressMaximum}</progress><div className="metrics"><Metric label="Discovered" value={crawl.discovered}/><Metric label="Queued" value={crawl.queued}/><Metric label="Fetched" value={crawl.fetched}/><Metric label="Analysed" value={crawl.analysed}/><Metric label="Failed" value={crawl.failed}/></div>{!terminal.has(crawl.status) && <button className="secondary danger" onClick={cancel}>Cancel crawl</button>}</div></section>}
    {summary && <section className="results" aria-labelledby="results-heading"><div className="section-heading"><p className="kicker">Audit evidence</p><h2 id="results-heading">Results</h2><p>Findings are technical observations with stored rule versions and evidence—not guarantees of indexing or rankings.</p></div><div className="summary-grid"><Metric label="Errors" value={summary.issues_by_severity.error ?? 0}/><Metric label="Warnings" value={summary.issues_by_severity.warning ?? 0}/><Metric label="Information" value={summary.issues_by_severity.info ?? 0}/><Metric label="Analysed pages" value={summary.Analysed}/></div><div className="tabs" role="tablist" aria-label="Audit result type"><button role="tab" aria-selected={view === "issues"} onClick={() => setView("issues")}>Issues</button><button role="tab" aria-selected={view === "pages"} onClick={() => setView("pages")}>Pages</button></div>{view === "issues" ? <IssueTable rows={issueRows}/> : <PageTable rows={pageRows}/>}</section>}
  </main>;
}

function Metric({label,value}:{label:string;value:number}) { return <div className="metric"><span>{label}</span><strong>{value.toLocaleString()}</strong></div>; }
function IssueTable({rows}:{rows:IssueRecord[]}) { return <div className="table-wrap"><table><thead><tr><th>Severity</th><th>Rule</th><th>Evidence</th></tr></thead><tbody>{rows.map((row)=><tr key={row.id}><td><span className={`severity ${row.severity}`}>{row.severity}</span></td><td className="mono">{row.rule_id}</td><td><code>{row.EvidenceJSON}</code></td></tr>)}</tbody></table>{rows.length===0&&<p className="empty">No issues in this result page.</p>}</div>; }
function PageTable({rows}:{rows:PageRecord[]}) { return <div className="table-wrap"><table><thead><tr><th>Status</th><th>URL</th><th>Title</th><th>Depth</th></tr></thead><tbody>{rows.map((row)=><tr key={row.id}><td>{row.StatusCode}</td><td><span className="url-cell">{row.URL}</span></td><td>{row.Title || <span className="muted">Missing</span>}</td><td>{row.Depth}</td></tr>)}</tbody></table>{rows.length===0&&<p className="empty">No analysed pages.</p>}</div>; }
