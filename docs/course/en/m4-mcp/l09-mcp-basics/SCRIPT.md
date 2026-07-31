# Lesson 9 — What MCP Does and Why It Matters

**Target duration:** 4 minutes

## Narration

[On screen: AI client ↔ MCP server ↔ SEO Screaming Toad.]

MCP means Model Context Protocol. It is an open standard that allows AI applications to connect with external data, tools, and workflows. The official documentation compares it to a USB-C connection for AI applications: a common interface instead of a unique integration for every combination.

Without MCP, a user may need to understand API routes, authentication, JSON schemas, crawl IDs, pagination, and report formats. Through MCP, an AI agent receives clear tools such as `project_create`, `crawl_preview_scope`, `crawl_start`, `crawl_status`, `issue_list`, `issue_explain`, and `report_export`.

This makes a technical crawler conversational. A user can ask for a bounded audit, request evidence for the highest-priority findings, and export a report without manually operating every table and API call.

[On screen: safety boundary—no generic shell, arbitrary SQL, unrestricted fetch, or arbitrary output path.]

Convenience must preserve control. Toad's MCP server communicates with the AI client over standard input and output, then calls the authenticated crawler API on local loopback. It does not open a public MCP network port.

The server exposes specialized SEO operations rather than a generic shell, arbitrary SQL, unrestricted filesystem access, or an unrestricted HTTP fetcher. An MCP caller cannot disable TLS verification, robots enforcement, target guards, redirect validation, or response budgets.

Starting a crawl still contacts an external website, so the user must authorize the target and scope. MCP reduces technical barriers; it does not remove permission, judgment, or responsibility.
