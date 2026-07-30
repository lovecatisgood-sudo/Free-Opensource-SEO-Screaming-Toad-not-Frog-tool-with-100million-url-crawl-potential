# SEO Screaming Toad MCP server

`seo-auditor-mcp` lets an AI agent operate the local SEO Screaming Toad application through Model Context Protocol (MCP). It uses the official Go MCP SDK and communicates over stdio. It does not open an MCP network port.

## Architecture

```text
AI agent / MCP client
        │ stdio
        ▼
seo-auditor-mcp
        │ authenticated loopback HTTP
        ▼
SEO Screaming Toad local supervisor
        │
        ├── guarded crawler and renderer
        ├── SQLite/WAL state
        └── managed report artifacts
```

Start the main `seo-auditor` application before connecting the MCP server. Both processes must use the same numeric loopback host and port. The defaults are `127.0.0.1:7331`.

## Client configuration

Build or download both `seo-auditor` and `seo-auditor-mcp`. Then add an MCP stdio entry to the AI client. A generic JSON example is stored at [`examples/mcp-client.json`](../examples/mcp-client.json).

```json
{
  "mcpServers": {
    "seo-screaming-toad": {
      "command": "/absolute/path/to/seo-auditor-mcp",
      "env": {
        "SEO_AUDITOR_BIND_HOST": "127.0.0.1",
        "SEO_AUDITOR_BIND_PORT": "7331"
      }
    }
  }
}
```

Use `seo-auditor-mcp.exe` on Windows. Configuration file names differ between MCP clients, but the command and environment values are the same. Standard output is reserved for MCP protocol frames; server errors are written to standard error.

After connection, call `health_get`. A fully connected response includes `"status":"ready"` and `"api_connected":true`. `adapter_ready` means the protocol adapter is running without a configured API caller, which is useful only in tests.

## Tool catalog

### Projects and profiles

- `project_create` — create a local project.
- `project_list` — list projects with opaque-cursor pagination.
- `profile_create` — create a safe profile from practical crawl settings.
- `profile_list` — list reusable profiles and their complete stored configurations.
- `crawl_preview_scope` — normalize candidate URLs and explain scope decisions without fetching them.

`profile_create` applies conservative defaults. `maximum_urls` defaults to 10,000, `maximum_depth` to 50, `rendering_mode` to `raw`, and `response_compression` to `gzip`.

### Crawl lifecycle

- `crawl_start` — start a stored profile and return immediately with a crawl ID.
- `crawl_status` — read counters, state and terminal reason.
- `crawl_pause`, `crawl_resume`, `crawl_cancel` — explicit lifecycle mutations.
- `crawl_list` — list project crawl history.
- `crawl_timeline` — list persisted lifecycle events.

Every `crawl_start` requires a caller-generated `idempotency_key` from 1 to 128 characters. Repeating the same key for the same project and profile in one MCP process returns the original start result instead of creating a second crawl.

### Audit evidence

- `audit_summary` — response, rendering and issue distributions.
- `issue_list` — filtered, paginated findings.
- `issue_explain` — versioned guidance, evidence and limitations for one finding.
- `page_list` — searchable, paginated crawled pages.
- `page_get` — bounded raw/rendered page, relationship and issue evidence.
- `link_list` — paginated crawl-graph links.
- `crawl_compare` — added, removed, changed, new and fixed results between crawls.

### Managed artifacts

- `report_export` — create CSV, NDJSON or XLSX reports in the application-owned artifact directory.
- `diagnostic_create` — create a metadata-only support artifact.
- `artifact_get` — retrieve artifact metadata and its approved managed path.

MCP tools never accept an arbitrary output path.

## Recommended agent workflow

1. Call `health_get`.
2. Use `project_list`; call `project_create` only when the user wants a new project.
3. Use `profile_list`. If needed, create a profile with a public seed URL and conservative limits.
4. Copy the stored configuration returned by `profile_list` into `crawl_preview_scope` and present important exclusions or limits.
5. After explicit operator intent, call `crawl_start` once with a unique idempotency key.
6. Poll `crawl_status` every 2–5 seconds. Stop polling at `completed`, `cancelled`, `failed`, or `limit_reached`.
7. Read `audit_summary`, then paginate `issue_list` and `page_list` rather than requesting unbounded output.
8. Use `issue_explain` and representative `page_get` evidence before recommending changes.
9. Export a managed report only when requested.

Example profile input:

```json
{
  "project_id": "project_…",
  "name": "Public site — raw HTML",
  "seed_url": "https://example.com/",
  "maximum_urls": 10000,
  "maximum_depth": 50,
  "allow_subdomains": false,
  "rendering_mode": "raw",
  "response_compression": "gzip",
  "exclude_path_regex": ["^/account", "^/checkout"]
}
```

## Security boundary

The MCP server exposes no generic shell, SQL, filesystem, browser or HTTP-fetch primitive. It cannot disable TLS verification, robots enforcement, target guards, redirect validation, response budgets or private-address rejection. It cannot supply arbitrary request headers, credentials, renderer executables or output paths.

Starting and resuming crawls can contact authorized public targets and are marked as open-world mutations in MCP tool annotations. Read-only tools are explicitly annotated so compatible clients can distinguish inspection from mutation.

## Troubleshooting

- **`local crawler API is unavailable`** — start `seo-auditor` and confirm both processes use the same `SEO_AUDITOR_BIND_HOST` and `SEO_AUDITOR_BIND_PORT`.
- **Profile validation error** — check the public seed URL, URL/depth limits, rendering mode and regular expressions.
- **Rendered mode unavailable** — the main application must be started with its trusted renderer configuration; MCP cannot set renderer paths.
- **Protocol parse errors** — do not write logs or wrapper output to the MCP process's standard output.
- **Crawl stops at `limit_reached`** — inspect `terminal_reason`; this is an incomplete bounded crawl, not a completed audit.
