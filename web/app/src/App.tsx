export interface BuildStatus {
  phase: string;
  detail: string;
}

export const foundationStatus: readonly BuildStatus[] = [
  { phase: "Guarded network", detail: "Policy boundary active" },
  { phase: "Durable storage", detail: "SQLite/WAL active" },
  { phase: "MCP", detail: "stdio transport active" },
] as const;

export function App() {
  return (
    <main className="shell">
      <header className="hero">
        <p className="eyebrow">Local-first technical SEO</p>
        <h1>SEO Auditor</h1>
        <p className="subtitle">A secure, durable crawler built for trustworthy audits.</p>
      </header>
      <section aria-labelledby="foundation-heading" className="panel">
        <div>
          <p className="section-label">Build status</p>
          <h2 id="foundation-heading">Foundation online</h2>
        </div>
        <ul className="status-grid">
          {foundationStatus.map((status) => (
            <li key={status.phase}>
              <span className="status-dot" aria-hidden="true" />
              <div>
                <strong>{status.phase}</strong>
                <span>{status.detail}</span>
              </div>
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}

