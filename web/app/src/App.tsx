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
  const [schedules, setSchedules] = useState<api.ScheduledAudit[]>([]);
  const [scheduleProfileID, setScheduleProfileID] = useState("");
  const [scheduleName, setScheduleName] = useState("Weekly audit");
  const [scheduleInterval, setScheduleInterval] = useState(604800);
  const [profileName, setProfileName] = useState("Default profile");
  const [seed, setSeed] = useState("");
  const [maximumURLs, setMaximumURLs] = useState(10_000);
  const [allowSubdomains, setAllowSubdomains] = useState(false);
  const [renderingMode, setRenderingMode] = useState<"raw" | "rendered">("raw");
  const [retainDOM, setRetainDOM] = useState(false);
  const [captureScreenshot, setCaptureScreenshot] = useState(false);
  const [runAccessibility, setRunAccessibility] = useState(false);
  const [responseCompression, setResponseCompression] = useState<"gzip" | "disabled">("gzip");
  const [excludePath, setExcludePath] = useState("");
  const [authenticationMode, setAuthenticationMode] = useState<"none" | "bearer" | "basic" | "cookie">("none");
  const [authenticationUser, setAuthenticationUser] = useState("");
  const [preview, setPreview] = useState<api.ScopeDecision[]>([]);
  const [crawl, setCrawl] = useState<api.CrawlProgress | null>(null);
  const [summary, setSummary] = useState<api.Summary | null>(null);
  const [pageRows, setPageRows] = useState<api.PageRecord[]>([]);
  const [issueRows, setIssueRows] = useState<api.IssueRecord[]>([]);
  const [pageCursor, setPageCursor] = useState("");
  const [issueCursor, setIssueCursor] = useState("");
  const [search, setSearch] = useState("");
  const [view, setView] = useState<"issues" | "pages" | "custom" | "architecture">("issues");
	const [architectureGraph, setArchitectureGraph] = useState<api.ArchitectureGraph | null>(null);
  const [detail, setDetail] = useState<api.PageDetail | null>(null);
  const [explanation, setExplanation] = useState<api.IssueExplanation | null>(null);
  const [comparison, setComparison] = useState<api.Comparison | null>(null);
  const [baseID, setBaseID] = useState("");
  const [targetID, setTargetID] = useState("");
  const [artifactMessage, setArtifactMessage] = useState("");
  const [customAudits, setCustomAudits] = useState<api.CustomAuditDefinitionRecord[]>([]);
  const [customRows, setCustomRows] = useState<api.CustomAuditResult[]>([]);
  const [customCursor, setCustomCursor] = useState("");
  const [customID, setCustomID] = useState("extract-headings");
  const [customName, setCustomName] = useState("Extract headings");
  const [customSelector, setCustomSelector] = useState("h1");
  const [customSelectorKind, setCustomSelectorKind] = useState<"css" | "xpath">("css");
  const [customMode, setCustomMode] = useState<"raw" | "rendered">("raw");
  const [customExtraction, setCustomExtraction] = useState<"text" | "html" | "attribute" | "count">("text");
  const [customAttribute, setCustomAttribute] = useState("");
  const [customCondition, setCustomCondition] = useState<"always" | "exists" | "absent" | "equals" | "contains" | "regex">("exists");
  const [customPattern, setCustomPattern] = useState("");
  const [customMessage, setCustomMessage] = useState("");
  const [customPreviewHTML, setCustomPreviewHTML] = useState("<main><h1>Example heading</h1></main>");
  const [customPreviewResult, setCustomPreviewResult] = useState("");
  const [secretStoreAvailable, setSecretStoreAvailable] = useState<boolean | null>(null);
  const [credentialReference, setCredentialReference] = useState("secret_google");
  const [credentialValue, setCredentialValue] = useState("");
  const [integrationProvider, setIntegrationProvider] = useState<"pagespeed" | "crux" | "search-console" | "ga4">("pagespeed");
  const [integrationTarget, setIntegrationTarget] = useState("");
  const [integrationStrategy, setIntegrationStrategy] = useState<"mobile" | "desktop">("mobile");
  const [integrationStartDate, setIntegrationStartDate] = useState("");
  const [integrationEndDate, setIntegrationEndDate] = useState("");
  const [ga4Property, setGA4Property] = useState("");
  const [integrationRows, setIntegrationRows] = useState<api.IntegrationObservation[]>([]);

  const configuration = (): api.CrawlConfiguration => ({
    seed_url: seed,
    allowed_hosts: [],
    allow_subdomains: allowSubdomains,
    exclude_path_regex: excludePath ? [excludePath] : [],
    rendering_mode: renderingMode,
    response_compression: responseCompression,
    rendered_evidence: { retain_dom: retainDOM, capture_screenshot: captureScreenshot, run_accessibility: runAccessibility, maximum_page_bytes: 8_388_608, maximum_crawl_bytes: 1_073_741_824, retention_days: 7 },
    authentication: { mode: authenticationMode, ...(authenticationMode !== "none" ? { credential_reference: credentialReference } : {}), ...(authenticationMode === "basic" ? { username: authenticationUser } : {}) },
    limits: { ...defaultLimits, maximum_urls: maximumURLs },
  });
  const customDefinition = (): api.CustomAuditDefinition => ({ schema_version: 1, id: customID, name: customName, enabled: true, mode: customMode, selector_kind: customSelectorKind, selector: customSelector, extraction: { kind: customExtraction, ...(customExtraction === "attribute" ? { attribute: customAttribute } : {}) }, condition: { kind: customCondition, ...(["equals", "contains", "regex"].includes(customCondition) ? { pattern: customPattern } : {}) }, ...(customMessage ? { finding: { severity: "warning" as const, message: customMessage } } : {}), limits: { maximum_matches: 100, maximum_value_bytes: 4096, maximum_total_bytes: 65536 } });
  const fail = (reason: unknown, fallback: string) => setError(reason instanceof Error ? reason.message : fallback);
  const refreshProjects = async () => {
    const result = await api.listProjects();
    setProjects(result.items);
  };
  const selectProject = async (project: api.Project) => {
    setSelectedProject(project);
    setError("");
    const [profilePage, crawlPage, definitions, observations, schedulePage] = await Promise.all([api.listProfiles(project.project_id), api.listCrawls(project.project_id), api.listCustomAudits(project.project_id), api.integrationObservations(project.project_id), api.listSchedules(project.project_id)]);
    setProfiles(profilePage.items);
    setHistory(crawlPage.items);
    setCustomAudits(definitions);
    setIntegrationRows(observations.items);
    setSchedules(schedulePage.items);
    if (!profilePage.items.some((profile) => profile.profile_id === scheduleProfileID)) setScheduleProfileID(profilePage.items[0]?.profile_id ?? "");
  };

  useEffect(() => {
    bootstrap();
    async function bootstrap() {
      try {
        await api.bootstrap();
        try { setSecretStoreAvailable((await api.secretStoreStatus()).available); } catch { setSecretStoreAvailable(false); }
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
    const [nextSummary, nextPages, nextIssues, nextCustom] = await Promise.all([api.auditSummary(id), api.pages(id, query), api.issues(id, query), api.customAuditResults(id)]);
    setSummary(nextSummary);
    setPageRows(nextPages.items);
    setIssueRows(nextIssues.items);
    setCustomRows(nextCustom.items);
    setPageCursor(nextPages.next_cursor ?? "");
    setIssueCursor(nextIssues.next_cursor ?? "");
    setCustomCursor(nextCustom.next_cursor ?? "");
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
    if (view === "custom" && customCursor) {
      const result = await api.customAuditResults(crawl.crawl_id, customCursor);
      setCustomRows((rows) => [...rows, ...result.items]);
      setCustomCursor(result.next_cursor ?? "");
    }
  }
	async function showArchitecture() {
		if (!crawl) return;
		setView("architecture");
		try { setArchitectureGraph(await api.architecture(crawl.crawl_id)); }
		catch (reason) { fail(reason, "Could not load architecture graph"); }
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
  async function saveCustomAudit(event: FormEvent) {
    event.preventDefault();
    if (!selectedProject) return;
    setBusy(true);
    try {
      await api.putCustomAudit(selectedProject.project_id, customDefinition());
      setCustomAudits(await api.listCustomAudits(selectedProject.project_id));
    } catch (reason) { fail(reason, "Could not save custom audit"); } finally { setBusy(false); }
  }
  async function previewCustomAudit() {
    try { setCustomPreviewResult(JSON.stringify(await api.previewCustomAudit(customDefinition(), customPreviewHTML), null, 2)); }
    catch (reason) { fail(reason, "Could not preview custom audit"); }
  }
  async function removeCustomAudit(id: string) {
    if (!selectedProject) return;
    try { await api.deleteCustomAudit(selectedProject.project_id, id); setCustomAudits(await api.listCustomAudits(selectedProject.project_id)); }
    catch (reason) { fail(reason, "Could not delete custom audit"); }
  }
  async function storeCredential(event: FormEvent) {
    event.preventDefault();
    try {
      await api.putSecret(credentialReference, credentialValue);
      setCredentialValue("");
      setSecretStoreAvailable(true);
    } catch (reason) { fail(reason, "Could not store credential securely"); }
  }
  async function revokeCredential() {
    try { await api.deleteSecret(credentialReference); setCredentialValue(""); }
    catch (reason) { fail(reason, "Could not revoke credential"); }
  }
  async function runIntegration(event: FormEvent) {
    event.preventDefault();
    if (!selectedProject) return;
    setBusy(true);
    try {
      if (integrationProvider === "pagespeed") await api.runPageSpeed(selectedProject.project_id, { credential_reference: secretStoreAvailable ? credentialReference : "", target: integrationTarget, strategy: integrationStrategy });
      if (integrationProvider === "crux") await api.runCrUX(selectedProject.project_id, { credential_reference: credentialReference, request: { url: integrationTarget } });
      if (integrationProvider === "search-console") await api.runSearchConsole(selectedProject.project_id, { credential_reference: credentialReference, request: { site_url: integrationTarget, start_date: integrationStartDate, end_date: integrationEndDate, dimensions: ["page"], row_limit: 1000, data_state: "final" } });
      if (integrationProvider === "ga4") await api.runGA4(selectedProject.project_id, { credential_reference: credentialReference, request: { property_id: ga4Property, start_date: integrationStartDate, end_date: integrationEndDate, limit: 1000 } });
      setIntegrationRows((await api.integrationObservations(selectedProject.project_id)).items);
    } catch (reason) { fail(reason, "Could not run external integration"); } finally { setBusy(false); }
  }
  async function createSchedule(event: FormEvent) {
    event.preventDefault(); if (!selectedProject) return;
    try { await api.createSchedule(selectedProject.project_id,{profile_id:scheduleProfileID,name:scheduleName,interval_seconds:scheduleInterval});setSchedules((await api.listSchedules(selectedProject.project_id)).items); }
    catch(reason){fail(reason,"Could not create scheduled audit");}
  }
  async function removeSchedule(id:string){if(!selectedProject)return;try{await api.deleteSchedule(selectedProject.project_id,id);setSchedules((await api.listSchedules(selectedProject.project_id)).items);}catch(reason){fail(reason,"Could not delete scheduled audit");}}

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
        <span>17 versioned audit families</span>
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
                {renderingMode === "rendered" && <><label className="check"><input type="checkbox" checked={retainDOM} onChange={(event) => setRetainDOM(event.target.checked)} /><span>Retain redacted DOM (7 days)</span></label><label className="check"><input type="checkbox" checked={captureScreenshot} onChange={(event) => setCaptureScreenshot(event.target.checked)} /><span>Capture viewport screenshot</span></label><label className="check"><input type="checkbox" checked={runAccessibility} onChange={(event) => setRunAccessibility(event.target.checked)} /><span>Run axe accessibility checks</span></label></>}
                <label>Compression<select value={responseCompression} onChange={(event) => setResponseCompression(event.target.value as "gzip" | "disabled")}><option value="gzip">Gzip (default)</option><option value="disabled">Disabled</option></select></label>
                <label>Authentication<select value={authenticationMode} onChange={(event) => { setAuthenticationMode(event.target.value as typeof authenticationMode); if (event.target.value !== "none") setRenderingMode("raw"); }}><option value="none">None</option><option value="bearer">Bearer token reference</option><option value="basic">Basic credentials reference</option><option value="cookie">Cookie reference</option></select></label>
                {authenticationMode !== "none" && <label>Secret reference<input required pattern="secret_[A-Za-z0-9][A-Za-z0-9._-]{0,99}" value={credentialReference} onChange={(event) => setCredentialReference(event.target.value)} /></label>}
                {authenticationMode === "basic" && <label>Username<input required maxLength={256} value={authenticationUser} onChange={(event) => setAuthenticationUser(event.target.value)} /></label>}
                <label className="exclude-field">Exclude path regex<input placeholder="^/private" value={excludePath} onChange={(event) => setExcludePath(event.target.value)} /></label>
                <label className="check"><input type="checkbox" checked={allowSubdomains} onChange={(event) => setAllowSubdomains(event.target.checked)} /><span>Include subdomains</span></label>
                <div className="actions"><button type="button" className="secondary" onClick={previewScope}>Preview scope</button><button className="primary" disabled={busy}>Save profile</button></div>
              </form>
              <p className="form-footnote">Robots, DNS/IP, redirect, response-size and rate controls remain enforced. Retained DOM is redacted; screenshots are opt-in because page pixels may contain personal data.</p>
              {preview.length > 0 && <div className="scope-preview" aria-live="polite">{preview.map((item) => <p key={item.url} className={item.allowed ? "allowed" : "blocked"}><strong>{item.allowed ? "Allowed" : "Excluded"}</strong> {item.normalized_url ?? item.url} {item.reason && <span>— {item.reason}</span>}</p>)}</div>}
              <div className="profile-grid">
                {profiles.map((profile) => <article className="profile-card" key={profile.profile_id}><div><strong>{profile.name}</strong><span>v{profile.version} · {profile.configuration.rendering_mode} · {profile.configuration.response_compression || "gzip"}</span><span>{profile.configuration.limits.maximum_urls.toLocaleString()} URL ceiling</span></div><button className="run-button" disabled={busy} onClick={() => startProfile(profile)}><span>▶</span> Start audit</button></article>)}
                {profiles.length === 0 && <p className="empty">Save this configuration to create the first reusable crawl profile.</p>}
              </div>
              {profiles.length > 0 && <div className="schedule-workbench"><form onSubmit={createSchedule} className="compare-form"><label>Schedule name<input required maxLength={200} value={scheduleName} onChange={(event)=>setScheduleName(event.target.value)}/></label><label>Profile<select required value={scheduleProfileID} onChange={(event)=>setScheduleProfileID(event.target.value)}>{profiles.map((profile)=><option key={profile.profile_id} value={profile.profile_id}>{profile.name}</option>)}</select></label><label>Frequency<select value={scheduleInterval} onChange={(event)=>setScheduleInterval(Number(event.target.value))}><option value={86400}>Daily</option><option value={604800}>Weekly</option><option value={2592000}>Every 30 days</option></select></label><button className="secondary">Schedule audit</button></form><div className="profile-grid">{schedules.map((item)=><article className="profile-card" key={item.schedule_id}><div><strong>{item.name}</strong><span>Next: {new Date(item.next_run_at).toLocaleString()}</span><span>{item.last_error||item.last_crawl_id||"Not run yet"}</span></div><button className="secondary danger" onClick={()=>removeSchedule(item.schedule_id)}>Delete</button></article>)}</div><p className="form-footnote">Schedules run only while the local application is open. Each run reuses the stored guarded profile and will not overlap an active run from the same schedule.</p></div>}
            </section>

            <section className="panel integration-panel" aria-labelledby="integration-heading">
              <div className="panel-toolbar"><PanelTitle eyebrow="Optional external evidence" title="Performance & search integrations" id="integration-heading" /><span className="toolbar-chip">{secretStoreAvailable ? "OS credential store ready" : "Credential store unavailable"}</span></div>
              <p className="form-footnote">These actions explicitly send the selected URL, origin, Search Console property, or GA4 property to Google. Results are stored separately as lab, field, or external API evidence. Credential values stay in the operating-system secret store and are never returned by the API or MCP.</p>
              <form onSubmit={storeCredential} className="crawl-form integration-form">
                <label>Credential reference<input required pattern="secret_[A-Za-z0-9][A-Za-z0-9._-]{0,99}" value={credentialReference} onChange={(event) => setCredentialReference(event.target.value)} /></label>
                <label className="seed-field">Credential value<input type="password" required autoComplete="off" value={credentialValue} onChange={(event) => setCredentialValue(event.target.value)} placeholder="API key, access token, or OAuth refresh JSON" /></label>
                <div className="actions"><button className="primary" disabled={!secretStoreAvailable || !credentialValue}>Store securely</button><button type="button" className="secondary" disabled={!secretStoreAvailable} onClick={revokeCredential}>Revoke reference</button></div>
              </form>
              <form onSubmit={runIntegration} className="crawl-form integration-form">
                <label>Provider<select value={integrationProvider} onChange={(event) => setIntegrationProvider(event.target.value as typeof integrationProvider)}><option value="pagespeed">PageSpeed Insights (lab)</option><option value="crux">Chrome UX Report (field)</option><option value="search-console">Google Search Console</option><option value="ga4">Google Analytics 4</option></select></label>
                {integrationProvider !== "ga4" && <label className="seed-field">{integrationProvider === "search-console" ? "Property URL" : "Public URL"}<input type={integrationProvider === "search-console" ? "text" : "url"} required value={integrationTarget} onChange={(event) => setIntegrationTarget(event.target.value)} placeholder={integrationProvider === "search-console" ? "https://example.com/ or sc-domain:example.com" : "https://example.com/"} /></label>}
                {integrationProvider === "pagespeed" && <label>Strategy<select value={integrationStrategy} onChange={(event) => setIntegrationStrategy(event.target.value as typeof integrationStrategy)}><option value="mobile">Mobile</option><option value="desktop">Desktop</option></select></label>}
                {(integrationProvider === "search-console" || integrationProvider === "ga4") && <><label>Start date<input type="date" required value={integrationStartDate} onChange={(event) => setIntegrationStartDate(event.target.value)} /></label><label>End date<input type="date" required value={integrationEndDate} onChange={(event) => setIntegrationEndDate(event.target.value)} /></label></>}
                {integrationProvider === "ga4" && <label>GA4 property ID<input required inputMode="numeric" pattern="[0-9]+" value={ga4Property} onChange={(event) => setGA4Property(event.target.value)} placeholder="123456789" /></label>}
                <div className="actions"><button className="primary" disabled={busy || (integrationProvider !== "pagespeed" && !secretStoreAvailable)}>Run explicit query</button></div>
              </form>
              <IntegrationTable rows={integrationRows} />
            </section>

            <section className="panel custom-audit-panel" aria-labelledby="custom-audit-heading">
              <div className="panel-toolbar"><PanelTitle eyebrow="Bounded expert workflows" title="Custom extraction & audits" id="custom-audit-heading" /><span className="toolbar-chip">No scripts · no network</span></div>
              <form onSubmit={saveCustomAudit} className="crawl-form custom-audit-form">
                <label>Stable ID<input required pattern="[A-Za-z0-9][A-Za-z0-9._-]{0,99}" value={customID} onChange={(event) => setCustomID(event.target.value)} /></label>
                <label>Name<input required maxLength={200} value={customName} onChange={(event) => setCustomName(event.target.value)} /></label>
                <label>Source<select value={customMode} onChange={(event) => setCustomMode(event.target.value as "raw" | "rendered")}><option value="raw">Raw HTML</option><option value="rendered">Rendered DOM</option></select></label>
                <label>Selector language<select value={customSelectorKind} onChange={(event) => setCustomSelectorKind(event.target.value as "css" | "xpath")}><option value="css">CSS subset</option><option value="xpath">XPath subset</option></select></label>
                <label className="seed-field">Selector<input required maxLength={512} value={customSelector} onChange={(event) => setCustomSelector(event.target.value)} placeholder={customSelectorKind === "css" ? "main article h2" : "//main//h2"} /></label>
                <label>Extract<select value={customExtraction} onChange={(event) => setCustomExtraction(event.target.value as typeof customExtraction)}><option value="text">Text</option><option value="html">Inner HTML</option><option value="attribute">Attribute</option><option value="count">Count</option></select></label>
                {customExtraction === "attribute" && <label>Attribute<input required value={customAttribute} onChange={(event) => setCustomAttribute(event.target.value)} placeholder="href" /></label>}
                <label>Condition<select value={customCondition} onChange={(event) => setCustomCondition(event.target.value as typeof customCondition)}><option value="always">Always when matched</option><option value="exists">Exists</option><option value="absent">Absent</option><option value="equals">Equals</option><option value="contains">Contains</option><option value="regex">Safe regex</option></select></label>
                {["equals", "contains", "regex"].includes(customCondition) && <label>Pattern<input maxLength={1024} value={customPattern} onChange={(event) => setCustomPattern(event.target.value)} /></label>}
                <label className="seed-field">Optional finding message<input maxLength={500} value={customMessage} onChange={(event) => setCustomMessage(event.target.value)} placeholder="Leave empty for extraction only" /></label>
                <div className="actions"><button className="primary" disabled={busy}>Save definition</button><button type="button" className="secondary" onClick={previewCustomAudit}>Preview locally</button></div>
                <label className="custom-preview-input">Preview HTML<textarea value={customPreviewHTML} onChange={(event) => setCustomPreviewHTML(event.target.value)} rows={4} /></label>
              </form>
              {customPreviewResult && <pre className="custom-preview" aria-live="polite">{customPreviewResult}</pre>}
              <div className="profile-grid">{customAudits.map((record) => <article className="profile-card" key={record.definition.id}><div><strong>{record.definition.name}</strong><span>{record.definition.selector_kind.toUpperCase()} · {record.definition.mode} · {record.definition.extraction.kind}</span><code>{record.definition.selector}</code></div><button className="secondary danger" onClick={() => removeCustomAudit(record.definition.id)}>Delete</button></article>)}{customAudits.length === 0 && <p className="empty">No custom definitions. Saved enabled definitions run automatically with this project's next crawl.</p>}</div>
              <p className="form-footnote">Definitions are declarative and limited to 100 matches, 4 KiB per value and 64 KiB total evidence by default.</p>
            </section>

            {crawl && <section className="panel progress-panel" aria-labelledby="crawl-progress-heading">
              <div className="campaign-title"><p className="eyebrow">Campaign status</p><h2 id="crawl-progress-heading">{crawl.status.replaceAll("_", " ")}</h2><p className="mono">{crawl.crawl_id}</p>{crawl.terminal_reason && <p>{crawl.terminal_reason.replaceAll("_", " ")}</p>}</div>
              <div className="progress-content"><progress max={progressMaximum} value={progressValue}>{progressValue} of {progressMaximum}</progress><div className="metrics"><Metric label="Discovered" value={crawl.discovered} /><Metric label="Queued" value={crawl.queued} /><Metric label="Fetched" value={crawl.fetched} /><Metric label="Analysed" value={crawl.analysed} /><Metric label="Failed" value={crawl.failed} /></div><div className="actions">{crawl.status === "running" && <button className="secondary" onClick={() => control("pause")}>Pause</button>}{crawl.status === "paused" && <button className="secondary" onClick={() => control("resume")}>Resume</button>}{!terminal.has(crawl.status) && <button className="secondary danger" onClick={() => control("cancel")}>Cancel</button>}</div></div>
            </section>}

            {summary && <section id="audit-results" className="panel results" aria-labelledby="results-heading">
              <div className="panel-toolbar"><PanelTitle eyebrow="Versioned audit evidence" title="Audit results" id="results-heading" /><span className="toolbar-chip">{summary.analysed.toLocaleString()} pages analysed</span></div>
              <div className="summary-grid"><Metric tone="error" label="Errors" value={summary.issues_by_severity.error ?? 0} /><Metric tone="warning" label="Warnings" value={summary.issues_by_severity.warning ?? 0} /><Metric label="Information" value={summary.issues_by_severity.info ?? 0} /><Metric label="Fetched" value={summary.fetched} />{Object.keys(summary.rendering_by_status ?? {}).length > 0 && <Metric label="Rendered" value={summary.rendering_by_status?.completed ?? 0} />}</div>
              <div className="result-tools"><form onSubmit={(event) => { event.preventDefault(); loadResults(crawl!.crawl_id).catch((reason) => fail(reason, "Search failed")); }}><label><span className="sr-only">Search current dataset</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Filter URL, title, rule or evidence…" /></label><button className="secondary">Search</button></form><button className="secondary" onClick={exportWorkbook}>Export XLSX</button><span className="artifact-message" aria-live="polite">{artifactMessage}</span></div>
              <div className="tabs" role="tablist" aria-label="Audit result type"><button role="tab" aria-selected={view === "issues"} onClick={() => setView("issues")}>Issues <span>{issueRows.length}</span></button><button role="tab" aria-selected={view === "pages"} onClick={() => setView("pages")}>Pages <span>{pageRows.length}</span></button><button role="tab" aria-selected={view === "custom"} onClick={() => setView("custom")}>Custom <span>{customRows.length}</span></button><button role="tab" aria-selected={view === "architecture"} onClick={showArchitecture}>Architecture <span>{architectureGraph?.nodes.length ?? 0}</span></button></div>
              {view === "issues" ? <IssueTable rows={issueRows} inspect={(row) => api.explainIssue(crawl!.crawl_id, row.id).then(setExplanation).catch((reason) => fail(reason, "Could not explain issue"))} /> : view === "pages" ? <PageTable rows={pageRows} inspect={(row) => api.pageDetail(crawl!.crawl_id, row.id).then(setDetail).catch((reason) => fail(reason, "Could not load page"))} /> : view === "custom" ? <CustomAuditTable rows={customRows} /> : <ArchitectureTable graph={architectureGraph} />}
              {view !== "architecture" && <button className="secondary load-more" disabled={view === "pages" ? !pageCursor : view === "issues" ? !issueCursor : !customCursor} onClick={loadMore}>Load next 100 rows</button>}
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

      <footer className="status-bar"><span><i className={ready ? "ready" : ""} /> {ready ? "Supervisor connected" : "Waiting for supervisor"}</span><span>SQLite/WAL persistence</span><span>MCP over stdio</span><span>Data stays local unless you export it</span><span className="creator-credit"><a href="https://github.com/lovecatisgood-sudo" target="_blank" rel="noreferrer">Siamese Cat Dev</a> from <a href="https://djai.academy" target="_blank" rel="noreferrer">DJAI Academy</a> &amp; With <a href="https://school.djai.academy" target="_blank" rel="noreferrer">DJAI Community</a></span><span className="repo-support"><a href="https://github.com/lovecatisgood-sudo/Free-Opensource-SEO-Screaming-Toad-not-Frog-tool-with-100million-url-crawl-potential" target="_blank" rel="noreferrer" aria-label="Support SEO Screaming Toad on GitHub">Support this project on GitHub ★</a></span></footer>

      {(detail || explanation) && <div className="drawer-backdrop" onClick={() => { setDetail(null); setExplanation(null); }}><aside className="drawer" aria-label="Evidence detail" onClick={(event) => event.stopPropagation()}><button className="drawer-close" aria-label="Close detail" onClick={() => { setDetail(null); setExplanation(null); }}>×</button>{detail && <><p className="eyebrow">Raw page evidence</p><h2>{detail.page.title || "Untitled page"}</h2><p className="mono break">{detail.page.url}</p><div className="summary-grid"><Metric label="Status" value={detail.page.status_code} /><Metric label="Depth" value={detail.page.depth} /><Metric label="Inlinks" value={detail.inlinks.length} /><Metric label="Outlinks" value={detail.outlinks.length} /></div><h3>Raw headings</h3>{detail.headings.map((heading, index) => <p key={index}>H{heading.level}: {heading.text}</p>)}{detail.rendered && <><p className="eyebrow drawer-section">Rendered evidence</p><h3>{detail.rendered.status}{detail.rendered.error_code ? ` — ${detail.rendered.error_code}` : ""}</h3><p>{detail.rendered.request_count.toLocaleString()} mediated requests · {detail.rendered.transferred_bytes.toLocaleString()} bytes · {detail.rendered.engine_version || "engine not reported"}</p><div className="summary-grid"><Metric label="Console" value={detail.rendered.console_count} /><Metric label="Failed resources" value={detail.rendered.resource_failure_count} /><Metric label="axe findings" value={detail.rendered.accessibility_count} /></div>{detail.artifacts.length > 0 && <p>{detail.artifacts.map((artifact) => <a className="artifact-link" key={artifact.artifact_id} href={`/api/v1/artifacts/${encodeURIComponent(artifact.artifact_id)}/download`}>{artifact.kind} ({Math.ceil(artifact.size_bytes / 1024)} KiB)</a>)}</p>}{detail.render_differences.length > 0 && <div className="table-wrap"><table><thead><tr><th>Field</th><th>Raw</th><th>Rendered</th></tr></thead><tbody>{detail.render_differences.map((item) => <tr key={item.field}><td>{item.field}</td><td>{item.raw_value}</td><td>{item.rendered_value}</td></tr>)}</tbody></table></div>}{detail.accessibility.length > 0 && <><h3>Automated accessibility findings</h3><p className="evidence-disclaimer">axe detects automatable conditions only; this is not a complete WCAG conformance assessment.</p>{detail.accessibility.map((item, index) => <article className="diagnostic-card" key={`${item.rule_id}-${index}`}><strong>{item.impact}: {item.rule_id}</strong><code>{item.target}</code><span>{item.help}</span></article>)}</>}{detail.resource_failures.length > 0 && <><h3>Failed or blocked resources</h3>{detail.resource_failures.map((item, index) => <p className="mono break" key={index}>{item.resource_type}: {item.url} — {item.error_code}</p>)}</>}</>}</>}{explanation && <><p className="eyebrow">{explanation.issue.rule_id} · version {explanation.rule.version}</p><h2>{explanation.rule.title}</h2><p><span className="evidence-label">{explanation.issue.classification}</span> from <span className="evidence-label">{explanation.issue.evidence_source}</span> evidence</p><p>{explanation.rule.remediation}</p><h3>Evidence</h3><pre>{JSON.stringify(JSON.parse(explanation.issue.evidence_json), null, 2)}</pre><h3>Limitations</h3><p>{explanation.rule.limitations}</p></>}</aside></div>}
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
  return <div className="table-wrap"><table><thead><tr><th>Severity</th><th>Classification</th><th>Source</th><th>Rule</th><th>Subject</th><th>Evidence</th><th /></tr></thead><tbody>{rows.map((row) => <tr key={row.id}><td><span className={`severity ${row.severity}`}>{row.severity}</span></td><td><span className="evidence-label">{row.classification}</span></td><td><span className="evidence-label">{row.evidence_source}</span></td><td className="mono">{row.rule_id}.v{row.rule_version}</td><td>{row.subject_type}</td><td><code>{row.evidence_json}</code></td><td><button className="text-button" onClick={() => inspect(row)}>Explain</button></td></tr>)}</tbody></table>{rows.length === 0 && <p className="empty">No matching issues.</p>}</div>;
}

function PageTable({ rows, inspect }: { rows: api.PageRecord[]; inspect: (row: api.PageRecord) => void }) {
  return <div className="table-wrap"><table><thead><tr><th>Status</th><th>URL</th><th>Raw title</th><th>Rendered title</th><th>Depth</th><th /></tr></thead><tbody>{rows.map((row) => <tr key={row.id}><td><span className={`http-status status-${Math.floor(row.status_code / 100)}`}>{row.status_code}</span></td><td><span className="url-cell">{row.url}</span></td><td>{row.title || <span className="muted">Missing</span>}</td><td>{row.render_status ? row.rendered_title || row.render_status : <span className="muted">Not requested</span>}</td><td>{row.depth}</td><td><button className="text-button" onClick={() => inspect(row)}>Inspect</button></td></tr>)}</tbody></table>{rows.length === 0 && <p className="empty">No matching pages.</p>}</div>;
}

function CustomAuditTable({ rows }: { rows: api.CustomAuditResult[] }) {
  return <div className="table-wrap"><table><thead><tr><th>Definition</th><th>Source</th><th>Page</th><th>Matches</th><th>Condition</th><th>Values</th></tr></thead><tbody>{rows.map((row) => <tr key={row.id}><td className="mono">{row.definition_id}.v{row.definition_schema_version}</td><td><span className="evidence-label">{row.mode}</span></td><td>{row.page_id}</td><td>{row.match_count}{row.truncated ? " (bounded)" : ""}</td><td>{row.finding ? <span className={`severity ${row.finding_severity ?? "warning"}`}>{row.finding_message ?? "Finding"}</span> : row.condition_met ? "Matched" : "Not matched"}</td><td><code>{JSON.stringify(row.values)}</code></td></tr>)}</tbody></table>{rows.length === 0 && <p className="empty">No custom-audit results for this crawl.</p>}</div>;
}

function IntegrationTable({ rows }: { rows: api.IntegrationObservation[] }) {
  return <div className="table-wrap integration-results"><table><thead><tr><th>Provider</th><th>Evidence</th><th>Scope</th><th>Freshness</th><th>Observed</th><th>Result</th></tr></thead><tbody>{rows.map((row) => <tr key={row.observation_id}><td>{row.provider}</td><td><span className="evidence-label">{row.evidence_source}</span></td><td className="mono break">{row.scope}</td><td>{row.freshness || "—"}</td><td>{row.observed_at}</td><td><code>{JSON.stringify(row.result)}</code></td></tr>)}</tbody></table>{rows.length === 0 && <p className="empty">No external observations stored for this project.</p>}</div>;
}

function ArchitectureTable({ graph }: { graph: api.ArchitectureGraph | null }) {
  if (!graph) return <p className="empty">Select Architecture to calculate the bounded internal-link view.</p>;
  return <><div className="architecture-summary"><span>{graph.nodes.length.toLocaleString()} nodes</span><span>{graph.edges.length.toLocaleString()} edges</span><span>{graph.nodes.filter((node) => node.orphan).length.toLocaleString()} orphans</span>{graph.truncated && <span>View bounded by configured limits</span>}</div><div className="table-wrap"><table><thead><tr><th>Score</th><th>Depth</th><th>Segment</th><th>Inlinks</th><th>Outlinks</th><th>Status</th><th>URL</th></tr></thead><tbody>{graph.nodes.map((node) => <tr key={node.url_id}><td>{node.score.toFixed(2)}</td><td>{node.depth}</td><td>{node.segment}</td><td>{node.inlinks}</td><td>{node.outlinks}</td><td>{node.status_code || "—"}</td><td><span className="url-cell">{node.url}</span>{node.orphan && <span className="severity warning">orphan</span>}</td></tr>)}</tbody></table></div></>;
}

function DJAIServiceRail() {
  return <aside className="service-rail" aria-label="DJAI services">
    <div className="djai-brand"><span>Built with</span><img src="/djai-logo.webp" alt="DJAI Academy" width="968" height="488" /></div>
    <p className="rail-intro">Need people behind the crawler? DJAI offers web, software and learning support.</p>
    <ServiceCard label="Web development" title="Launch a search-ready website" copy="Design, development and technical SEO foundations from one delivery team." href="https://www.djai.academy/web_promo/?lang=en" action="Explore web services" />
    <ServiceCard label="Software development" title="Finding a dev team that can deliver?" copy="Build your app with DJAI development today." href="https://www.djai.academy/portfolio/en/" action="Build with DJAI" />
    <ServiceCard label="Free online school" title="Learn vibe coding" copy="Join the DJAI online school community and learn to turn ideas into working software." href="https://school.djai.academy" action="Join the community" />
    <p className="external-note">Promotional links open external DJAI websites. They never change crawl findings.</p>
  </aside>;
}

function ServiceCard({ label, title, copy, href, action }: { label: string; title: string; copy: string; href: string; action: string }) {
  return <article className="service-card"><p>{label}</p><h2>{title}</h2><span>{copy}</span><a href={href} target="_blank" rel="noreferrer">{action} <b aria-hidden="true">↗</b></a></article>;
}
