# ADR-004: MCP over stdio with the official Go SDK

Status: Accepted  
Date: 2026-07-30

## Context

SEO Auditor needs an agent-facing interface without opening another network service or duplicating application logic.

## Decision

Use `github.com/modelcontextprotocol/go-sdk` and stdio transport. Tool handlers are thin adapters around application services. Tool inputs and outputs are typed and bounded. Mutating and read-only tools use distinct names. Standard output is reserved for MCP frames; diagnostics go to standard error.

The completed implementation exposes bounded project/profile setup, scope preview, crawl lifecycle, status/timeline, audit evidence, page/link inspection, comparisons and managed artifacts. Its initialize/list/call lifecycle and scripted audit workflow are tested through the SDK transport. Tool annotations distinguish read-only, additive, lifecycle and open-world operations.

## Consequences

- Network MCP is absent from version one.
- MCP cannot supply arbitrary paths, headers, commands, private-target exceptions or robots bypasses.
- Protocol and schema conformance tests are release gates.
