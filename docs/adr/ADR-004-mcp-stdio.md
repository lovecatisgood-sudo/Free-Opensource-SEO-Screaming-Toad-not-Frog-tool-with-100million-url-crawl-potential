# ADR-004: MCP over stdio with the official Go SDK

Status: Accepted  
Date: 2026-07-30

## Context

SEO Auditor needs an agent-facing interface without opening another network service or duplicating application logic.

## Decision

Use `github.com/modelcontextprotocol/go-sdk` and stdio transport. Tool handlers are thin adapters around application services. Tool inputs and outputs are typed and bounded. Mutating and read-only tools use distinct names. Standard output is reserved for MCP frames; diagnostics go to standard error.

The foundation implementation exposes only the read-only `health_get` tool. Its initialize/connect/call lifecycle is tested through the SDK's in-memory transport before the remaining PRD tools are added.

## Consequences

- Network MCP is absent from version one.
- MCP cannot supply arbitrary paths, headers, commands, private-target exceptions or robots bypasses.
- Protocol and schema conformance tests are release gates.

