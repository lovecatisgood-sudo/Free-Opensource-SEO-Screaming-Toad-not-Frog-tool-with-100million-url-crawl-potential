import { describe, expect, it } from "vitest";

describe("local application policy", () => { it("keeps the app outside public indexing", async () => { const html = await import("../index.html?raw"); expect(html.default).toContain('name="robots" content="noindex,nofollow,noarchive"'); }); });
