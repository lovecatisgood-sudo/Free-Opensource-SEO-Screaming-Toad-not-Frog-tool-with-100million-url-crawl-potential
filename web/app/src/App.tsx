import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import * as api from "./api";

const terminal = new Set(["completed", "cancelled", "failed", "limit_reached"]);
const defaultLimits: api.CrawlLimits = {
  maximum_urls: 10_000,
  maximum_depth: 50,
  maximum_duration: 86_400_000_000_000,
  maximum_body_bytes: 26_214_400,
  maximum_disk_bytes: 21_474_836_480,
  global_concurrency: 16,
  per_host_concurrency: 2,
  minimum_host_delay: 100_000_000,
};

export function App() {
  const [ready, setReady] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [projects, setProjects] = useState<api.Project[]>([]);
  const [selectedProject, setSelectedProject] = useState<api.Project | null>(null);
  const [projectName, setProjectName] = useState("");
  const [profiles, setProfiles] = useState<api.Profile[]>([]);
  const [history, setHistory] = useState<api.CrawlProgress[]>([]);
  const [profileName, setProfileName] = useState("Default profile");
  const [seed, setSeed] = useState("");
  const [maximumURLs, setMaximumURLs] = useState(10_000);
  const [allowSubdomains, setAllowSubdomains] = useState(false);
  const [renderingMode, setRenderingMode] = useState<"raw" | "rendered">("raw");
  const [responseCompression, setResponseCompression] = useState<"gzip" | "disabled">("gzip");
  const [excludePath, setExcludePath] = useState("");
  const [preview, setPreview] = useState<api.ScopeDecision[]>([]);
  const [crawl, setCrawl] = useState<api.CrawlProgress | null>(null);
  const [summary, setSummary] = useState<api.Summary | null>(null);
  const [pageRows, setPageRows] = useState<api.PageRecord[]>([]);
  const [issueRows, setIssueRows] = useState<api.IssueRecord[]>([]);
  const [pageCursor, setPageCursor] = useState("");
  const [issueCursor, setIssueCursor] = useState("");
  const [search, setSearch] = useState("");
  const [view, setView] = useState<"issues" | "pages">("issues");
  const [detail, setDetail] = useState<api.PageDetail | null>(null);
  const [explanation, setExplanation] = useState<api.IssueExplanation | null>(null);
  const [comparison, setComparison] = useState<api.Comparison | null>(null);
  const [baseID, setBaseID] = useState("");
  const [targetID, setTargetID] = useState("");
  const [artifactMessage, setArtifactMessage] = useState("");

  const configuration = (): api.CrawlConfiguration => ({
    seed_url: seed,
    allowed_hosts: [],
    allow_subdomains: allowSubdomains,
    exclude_path_regex: excludePath ? [excludePath] : [],
    rendering_mode: renderingMode,
    response_compression: responseCompression,
    limits: { ...defaultLimits, maximum_urls: maximumURLs },
  });
  const fail = (reason: unknown, fallback: string) => setError(reason instanceof Error ? reason.message : fallback);
  const refreshProjects = async () => {
    const result = await api.listProjects();
    setProjects(result.items);
  };
  const selectProject = async (project: api.Project) => {
    setSelectedProject(project);
    setError("");
    const [profilePage, crawlPage] = await Promise.all([api.listProfiles(project.project_id), api.listCrawls(project.project_id)]);
    setProfiles(profilePage.items);
    setHistory(crawlPage.items);
  };

  useEffect(() => {
    bootstrap();
    async function bootstrap() {
      try {
        await api.bootstrap();
        await refreshProjects();
        setReady(true);
      } catch (reason) {
        fail(reason, "Could not connect to the local API");
      }
    }
  }, []);
  useEffect(() => {
    if (!crawl || terminal.has(crawl.status) || crawl.status === "paused") return;
    const timer = window.setInterval(() => api.crawlStatus(crawl.crawl_id).then(setCrawl).catch((reason) => fail(reason, "Status request failed")), 1000);
    return () => window.clearInterval(timer);
  }, [crawl?.crawl_id, crawl?.status]);
  useEffect(() => {
    if (!crawl || !terminal.has(crawl.status)) return;
    loadResults(crawl.crawl_id).catch((reason) => fail(reason, "Could not load audit results"));
  }, [crawl?.crawl_id, crawl?.status]);

  async function createProject(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    try {
      const project = await api.createProject(projectName);
      await refreshProjects();
      await selectProject(project);
      setProjectName("");
    } catch (reason) {
      fail(reason, "Could not create project");
    } finally {
      setBusy(false);
    }
  }
  async function previewScope() {
    if (!selectedProject) return;
    try {
      const result = await api.previewScope(selectedProject.project_id, configuration(), [seed]);
      setPreview(result.decisions);
    } catch (reason) {
      fail(reason, "Could not preview scope");
    }
  }
  async function saveProfile(event: FormEvent) {
    event.preventDefault();
    if (!selectedProject) return;
    setBusy(true);
    try {
      await api.createProfile(selectedProject.project_id, profileName, configuration());
      setProfiles((await api.listProfiles(selectedProject.project_id)).items);
      setPreview([]);
    } catch (reason) {
      fail(reason, "Could not save profile");
    } finally {
      setBusy(false);
    }
  }
  async function startProfile(profile: api.Profile) {
    if (!selectedProject) return;
    setBusy(true);
    setSummary(null);
    setError("");
    try {
      const result = await api.startProfileCrawl(selectedProject.project_id, profile.profile_id);
      setCrawl(result.progress);
      setHistory((await api.listCrawls(selectedProject.project_id)).items);
    } catch (reason) {
      fail(reason, "Could not start crawl");
    } finally {
      setBusy(false);
    }
  }
  async function control(action: "pause" | "resume" | "cancel") {
    if (!crawl) return;
    try {
      await ({ pause: api.pauseCrawl, resume: api.resumeCrawl, cancel: api.cancelCrawl }[action])(crawl.crawl_id);
      setCrawl(await api.crawlStatus(crawl.crawl_id));
    } catch (reason) {
      fail(reason, `Could not ${action} crawl`);
    }
  }
  async function loadResults(id: string, query = search) {
    const [nextSummary, nextPages, nextIssues] = await Promise.all([api.auditSummary(id), api.pages(id, query), api.issues(id, query)]);
    setSummary(nextSummary);
    setPageRows(nextPages.items);
    setIssueRows(nextIssues.items);
    setPageCursor(nextPages.next_cursor ?? "");
    setIssueCursor(nextIssues.next_cursor ?? "");
  }
  async function loadMore() {
    if (!crawl) return;
    if (view === "pages" && pageCursor) {
      const result = await api.pages(crawl.crawl_id, search, pageCursor);
      setPageRows((rows) => [...rows, ...result.items]);
      setPageCursor(result.next_cursor ?? "");
    }
    if (view === "issues" && issueCursor) {
      const result = await api.issues(crawl.crawl_id, search, issueCursor);
      setIssueRows((rows) => [...rows, ...result.items]);
      setIssueCursor(result.next_cursor ?? "");
    }
  }
  async function runComparison() {
    try {
      setComparison(await api.compareCrawls(baseID, targetID));
    } catch (reason) {
      fail(reason, "Could not compare crawls");
    }
  }
  async function exportWorkbook() {
    if (!crawl) return;
    try {
      const artifact = await api.createExport(crawl.crawl_id, "workbook", "xlsx");
      setArtifactMessage(`Workbook ready: ${artifact.path}`);
    } catch (reason) {
      fail(reason, "Could not create report");
    }
  }

  const progressMaximum = Math.max(crawl?.discovered ?? 1, 1);
  const progressValue = crawl?.analysed ?? 0;

  return (
    <main className="app-shell">
      <header className="product-header">
        <div className="product-identity">
          <div className="toad-avatar"><img src="/screaming-toad.png" alt="SEO Screaming Toad mascot holding a magnifying glass" /></div>
          <div>
            <p className="product-overline">Free &amp; open-source technical SEO crawler</p>
            <h1>SEO Screaming Toad <span>— not Frog</span></h1>
            <p className="product-alias">DJAI Toad · local-first audit workbench</p>
          </div>
        </div>
        <div className="header-status">
          <span className="capacity-note">Segmented campaigns · theoretical 100M+ architecture</span>
          <span className={`connection ${ready ? "online" : ""}`}>{ready ? "Local API ready" : "Connecting…"}</span>
        </div>
      </header>

      <nav className="command-bar" aria-label="Workspace sections">
        <a href="#crawl-setup">Crawl setup</a>
        <a href="#audit-results">Audit results</a>
        <a href="#crawl-history">Crawl history</a>
        <span>13 versioned audit families</span>
        <span>Raw + rendered evidence</span>
        <span>Local data</span>
      </nav>

      {error && <div className="alert" role="alert"><span>{error}</span><button aria-label="Dismiss error" onClick={() => setError("")}>×</button></div>}

      <div className="workbench">
        <aside className="project-explorer" aria-label="Projects">
          <PanelTitle eyebrow="Workspace" title="Projects" />
          <form onSubmit={createProject} className="compact-form">
            <label>New project<input required maxLength={200} value={projectName} onChange={(event) => setProjectName(event.target.value)} placeholder="Client or website" /></label>
            <button className="primary" disabled={!ready || busy}>Create project</button>
          </form>
          <nav className="project-list" aria-label="Saved projects">
            {projects.map((project) => <button key={project.project_id} className={selectedProject?.project_id === project.project_id ? "project active" : "project"} onClick={() => selectProject(project).catch((reason) => fail(reason, "Could not load project"))}><span className="project-icon">▦</span><span><strong>{project.name}</strong><small>{project.archived ? "Archived" : "Active workspace"}</small></span></button>)}
            {projects.length === 0 && ready && <p className="empty">No projects yet. Create one to store profiles, crawls and comparisons.</p>}
          </nav>
          <div className="explorer-note"><strong>Evidence model</strong><span>Rule IDs, versions, remediation and limitations stay attached to every finding.</span></div>
        </aside>

        <div className="main-workspace">
          {selectedProject ? <>
            <section id="crawl-setup" className="panel setup-panel" aria-labelledby="profile-heading">
              <div className="panel-toolbar">
                <PanelTitle eyebrow={selectedProject.name} title="Crawl configuration" id="profile-heading" />
                <span className="toolbar-chip">Guarded fetch policy</span>
              </div>
              <form onSubmit={saveProfile} className="crawl-form">
                <label>Profile name<input required maxLength={200} value={profileName} onChange={(event) => setProfileName(event.target.value)} /></label>
                <label className="seed-field">Seed URL<input type="url" required placeholder="https://example.com/" value={seed} onChange={(event) => setSeed(event.target.value)} /></label>
                <label>URL ceiling<input type="number" min={1} max={100_000_000} value={maximumURLs} onChange={(event) => setMaximumURLs(Number(event.target.value))} /></label>
                <label>Rendering<select value={renderingMode} onChange={(event) => setRenderingMode(event.target.value as "raw" | "rendered")}><option value="raw">Raw HTML</option><option value="rendered">JavaScript rendered</option></select></label>
                <label>Compression<select value={responseCompression} onChange={(event) => setResponseCompression(event.target.value as "gzip" | "disabled")}><option value="gzip">Gzip (default)</option><option value="disabled">Disabled</option></select></label>
                <label className="exclude-field">Exclude path regex<input placeholder="^/private" value={excludePath} onChange={(event) => setExcludePath(event.target.value)} /></label>
                <label className="check"><input type="checkbox" checked={allowSubdomains} onChange={(event) => setAllowSubdomains(event.target.checked)} /><span>Include subdomains</span></label>
                <div className="actions"><button type="button" className="secondary" onClick={previewScope}>Preview scope</button><button className="primary" disabled={busy}>Save profile</button></div>
              </form>
              <p className="form-footnote">Robots policy, scope, DNS/IP, redirect, response-size and rate controls remain enforced for every profile.</p>
              {preview.length > 0 && <div className="scope-preview" aria-live="polite">{preview.map((item) => <p key={item.url} className={item.allowed ? "allowed" : "blocked"}><strong>{item.allowed ? "Allowed" : "Excluded"}</strong> {item.normalized_url ?? item.url} {item.reason && <span>— {item.reason}</span>}</p>)}</div>}
              <div className="profile-grid">
                {profiles.map((profile) => <article className="profile-card" key={profile.profile_id}><div><strong>{profile.name}</strong><span>v{profile.version} · {profile.configuration.rendering_mode} · {profile.configuration.response_compression || "gzip"}</span><span>{profile.configuration.limits.maximum_urls.toLocaleString()} URL ceiling</span></div><button className="run-button" disabled={busy} onClick={() => startProfile(profile)}><span>▶</span> Start audit</button></article>)}
                {profiles.length === 0 && <p className="empty">Save this configuration to create the first reusable crawl profile.</p>}
              </div>
            </section>

            {crawl && <section className="panel progress-panel" aria-labelledby="crawl-progress-heading">
              <div className="campaign-title"><p className="eyebrow">Campaign status</p><h2 id="crawl-progress-heading">{crawl.status.replaceAll("_", " ")}</h2><p className="mono">{crawl.crawl_id}</p>{crawl.terminal_reason && <p>{crawl.terminal_reason.replaceAll("_", " ")}</p>}</div>
              <div className="progress-content"><progress max={progressMaximum} value={progressValue}>{progressValue} of {progressMaximum}</progress><div className="metrics"><Metric label="Discovered" value={crawl.discovered} /><Metric label="Queued" value={crawl.queued} /><Metric label="Fetched" value={crawl.fetched} /><Metric label="Analysed" value={crawl.analysed} /><Metric label="Failed" value={crawl.failed} /></div><div className="actions">{crawl.status === "running" && <button className="secondary" onClick={() => control("pause")}>Pause</button>}{crawl.status === "paused" && <button className="secondary" onClick={() => control("resume")}>Resume</button>}{!terminal.has(crawl.status) && <button className="secondary danger" onClick={() => control("cancel")}>Cancel</button>}</div></div>
            </section>}

            {summary && <section id="audit-results" className="panel results" aria-labelledby="results-heading">
              <div className="panel-toolbar"><PanelTitle eyebrow="Versioned audit evidence" title="Audit results" id="results-heading" /><span className="toolbar-chip">{summary.analysed.toLocaleString()} pages analysed</span></div>
              <div className="summary-grid"><Metric tone="error" label="Errors" value={summary.issues_by_severity.error ?? 0} /><Metric tone="warning" label="Warnings" value={summary.issues_by_severity.warning ?? 0} /><Metric label="Information" value={summary.issues_by_severity.info ?? 0} /><Metric label="Fetched" value={summary.fetched} />{Object.keys(summary.rendering_by_status).length > 0 && <Metric label="Rendered" value={summary.rendering_by_status.completed ?? 0} />}</div>
              <div className="result-tools"><form onSubmit={(event) => { event.preventDefault(); loadResults(crawl!.crawl_id).catch((reason) => fail(reason, "Search failed")); }}><label><span className="sr-only">Search current dataset</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Filter URL, title, rule or evidence…" /></label><button className="secondary">Search</button></form><button className="secondary" onClick={exportWorkbook}>Export XLSX</button><span className="artifact-message" aria-live="polite">{artifactMessage}</span></div>
              <div className="tabs" role="tablist" aria-label="Audit result type"><button role="tab" aria-selected={view === "issues"} onClick={() => setView("issues")}>Issues <span>{issueRows.length}</span></button><button role="tab" aria-selected={view === "pages"} onClick={() => setView("pages")}>Pages <span>{pageRows.length}</span></button></div>
              {view === "issues" ? <IssueTable rows={issueRows} inspect={(row) => api.explainIssue(crawl!.crawl_id, row.id).then(setExplanation).catch((reason) => fail(reason, "Could not explain issue"))} /> : <PageTable rows={pageRows} inspect={(row) => api.pageDetail(crawl!.crawl_id, row.id).then(setDetail).catch((reason) => fail(reason, "Could not load page"))} />}
              <button className="secondary load-more" disabled={view === "pages" ? !pageCursor : !issueCursor} onClick={loadMore}>Load next 100 rows</button>
              <p className="evidence-disclaimer">Findings are technical observations with explicit limitations—not guarantees of indexing, rankings or rich-result eligibility.</p>
            </section>}

            <section id="crawl-history" className="panel history-panel">
              <PanelTitle eyebrow="Recrawl and compare" title="Crawl history" />
              <div className="history-list">{history.map((item) => <button key={item.crawl_id} onClick={() => setCrawl(item)}><span className={`status-dot ${item.status}`} /><span className="mono">{item.crawl_id}</span><strong>{item.status.replaceAll("_", " ")}</strong><span>{item.analysed.toLocaleString()} analysed</span><time>{new Date(item.updated_at).toLocaleDateString()}</time></button>)}{history.length === 0 && <p className="empty">Completed and interrupted crawls will appear here.</p>}</div>
              {history.length >= 2 && <div className="compare-form"><label>Base crawl<select value={baseID} onChange={(event) => setBaseID(event.target.value)}><option value="">Select crawl</option>{history.map((item) => <option key={item.crawl_id} value={item.crawl_id}>{item.crawl_id}</option>)}</select></label><label>Target crawl<select value={targetID} onChange={(event) => setTargetID(event.target.value)}><option value="">Select crawl</option>{history.map((item) => <option key={item.crawl_id} value={item.crawl_id}>{item.crawl_id}</option>)}</select></label><button className="secondary" disabled={!baseID || !targetID} onClick={runComparison}>Compare crawls</button></div>}
              {comparison && <div className="summary-grid comparison"><Metric label="Added" value={comparison.added_pages} /><Metric label="Removed" value={comparison.removed_pages} /><Metric label="Changed" value={comparison.changed_pages} /><Metric label="New issues" value={comparison.new_issues} /><Metric label="Fixed issues" value={comparison.fixed_issues} /><div className="metric"><span>Configuration</span><strong>{comparison.configuration_match ? "Match" : "Different"}</strong></div></div>}
            </section>
          </> : <section className="panel welcome">
            <img src="/screaming-toad.png" alt="SEO Screaming Toad product logo" width="1024" height="1024" />
            <div><p className="eyebrow">Start a technical audit</p><h2>Select or create a project</h2><p>Projects keep guarded crawl profiles, history, comparisons and evidence together. Configure a bounded public-site audit, run it locally, then inspect every finding down to its rule version and source evidence.</p><div className="welcome-steps"><span><b>1</b>Create project</span><span><b>2</b>Save profile</span><span><b>3</b>Run &amp; inspect</span></div></div>
          </section>}
        </div>

        <DJAIServiceRail />
      </div>

      <footer className="status-bar"><span><i className={ready ? "ready" : ""} /> {ready ? "Supervisor connected" : "Waiting for supervisor"}</span><span>SQLite/WAL persistence</span><span>MCP over stdio</span><span>Data stays local unless you export it</span></footer>

      {(detail || explanation) && <div className="drawer-backdrop" onClick={() => { setDetail(null); setExplanation(null); }}><aside className="drawer" aria-label="Evidence detail" onClick={(event) => event.stopPropagation()}><button className="drawer-close" aria-label="Close detail" onClick={() => { setDetail(null); setExplanation(null); }}>×</button>{detail && <><p className="eyebrow">Raw page evidence</p><h2>{detail.page.title || "Untitled page"}</h2><p className="mono break">{detail.page.url}</p><div className="summary-grid"><Metric label="Status" value={detail.page.status_code} /><Metric label="Depth" value={detail.page.depth} /><Metric label="Inlinks" value={detail.inlinks.length} /><Metric label="Outlinks" value={detail.outlinks.length} /></div><h3>Raw headings</h3>{detail.headings.map((heading, index) => <p key={index}>H{heading.level}: {heading.text}</p>)}{detail.rendered && <><p className="eyebrow drawer-section">Rendered evidence</p><h3>{detail.rendered.status}{detail.rendered.error_code ? ` — ${detail.rendered.error_code}` : ""}</h3><p>{detail.rendered.request_count.toLocaleString()} mediated requests · {detail.rendered.transferred_bytes.toLocaleString()} bytes</p>{detail.render_differences.length > 0 && <div className="table-wrap"><table><thead><tr><th>Field</th><th>Raw</th><th>Rendered</th></tr></thead><tbody>{detail.render_differences.map((item) => <tr key={item.field}><td>{item.field}</td><td>{item.raw_value}</td><td>{item.rendered_value}</td></tr>)}</tbody></table></div>}</>}</>}{explanation && <><p className="eyebrow">{explanation.issue.rule_id} · version {explanation.rule.version}</p><h2>{explanation.rule.title}</h2><p>{explanation.rule.remediation}</p><h3>Evidence</h3><pre>{JSON.stringify(JSON.parse(explanation.issue.evidence_json), null, 2)}</pre><h3>Limitations</h3><p>{explanation.rule.limitations}</p></>}</aside></div>}
    </main>
  );
}

function PanelTitle({ eyebrow, title, id }: { eyebrow: string; title: string; id?: string }) {
  return <div className="panel-title"><p className="eyebrow">{eyebrow}</p><h2 id={id}>{title}</h2></div>;
}

function Metric({ label, value, tone = "" }: { label: string; value: number; tone?: "" | "error" | "warning" }) {
  return <div className={`metric ${tone}`}><span>{label}</span><strong>{value.toLocaleString()}</strong></div>;
}

function IssueTable({ rows, inspect }: { rows: api.IssueRecord[]; inspect: (row: api.IssueRecord) => void }) {
  return <div className="table-wrap"><table><thead><tr><th>Severity</th><th>Rule</th><th>Subject</th><th>Evidence</th><th /></tr></thead><tbody>{rows.map((row) => <tr key={row.id}><td><span className={`severity ${row.severity}`}>{row.severity}</span></td><td className="mono">{row.rule_id}.v{row.rule_version}</td><td>{row.subject_type}</td><td><code>{row.evidence_json}</code></td><td><button className="text-button" onClick={() => inspect(row)}>Explain</button></td></tr>)}</tbody></table>{rows.length === 0 && <p className="empty">No matching issues.</p>}</div>;
}

function PageTable({ rows, inspect }: { rows: api.PageRecord[]; inspect: (row: api.PageRecord) => void }) {
  return <div className="table-wrap"><table><thead><tr><th>Status</th><th>URL</th><th>Raw title</th><th>Rendered title</th><th>Depth</th><th /></tr></thead><tbody>{rows.map((row) => <tr key={row.id}><td><span className={`http-status status-${Math.floor(row.status_code / 100)}`}>{row.status_code}</span></td><td><span className="url-cell">{row.url}</span></td><td>{row.title || <span className="muted">Missing</span>}</td><td>{row.render_status ? row.rendered_title || row.render_status : <span className="muted">Not requested</span>}</td><td>{row.depth}</td><td><button className="text-button" onClick={() => inspect(row)}>Inspect</button></td></tr>)}</tbody></table>{rows.length === 0 && <p className="empty">No matching pages.</p>}</div>;
}

function DJAIServiceRail() {
  return <aside className="service-rail" aria-label="DJAI services">
    <div className="djai-brand"><span>Built with</span><img src="/djai-logo.webp" alt="DJAI Academy" width="968" height="488" /></div>
    <p className="rail-intro">Need people behind the crawler? DJAI offers web, software and learning support.</p>
    <ServiceCard label="Web development" title="Launch a search-ready website" copy="Design, development and technical SEO foundations from one delivery team." href="https://djai.academy/webpromo/" action="Explore web services" />
    <ServiceCard label="Software development" title="Finding a dev team that can deliver?" copy="Build your app with DJAI development today." href="https://djai.academy/development" action="Build with DJAI" />
    <ServiceCard label="Free online school" title="Learn vibe coding" copy="Join the DJAI online school community and learn to turn ideas into working software." href="https://school.djai.academy" action="Join the community" />
    <p className="external-note">Promotional links open external DJAI websites. They never change crawl findings.</p>
  </aside>;
}

function ServiceCard({ label, title, copy, href, action }: { label: string; title: string; copy: string; href: string; action: string }) {
  return <article className="service-card"><p>{label}</p><h2>{title}</h2><span>{copy}</span><a href={href} target="_blank" rel="noreferrer">{action} <b aria-hidden="true">↗</b></a></article>;
}
