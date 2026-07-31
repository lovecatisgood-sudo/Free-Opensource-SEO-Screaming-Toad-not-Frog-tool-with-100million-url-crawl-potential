import { describe, expect, it } from "vitest";

describe("local application policy", () => {
  it("keeps the app outside public indexing", async () => {
    const html = await import("../index.html?raw");
    expect(html.default).toContain('name="robots" content="noindex,nofollow,noarchive"');
    expect(html.default).toContain("SEO Screaming Toad");
  });

  it("keeps capacity claims qualified and DJAI links explicit", async () => {
    const source = await import("./App.tsx?raw");
    expect(source.default).toContain("theoretical 100M+ architecture");
    expect(source.default).toContain("https://www.djai.academy/web_promo/?lang=en");
    expect(source.default).toContain("https://www.djai.academy/portfolio/en/");
    expect(source.default).toContain("https://school.djai.academy");
    expect(source.default).toContain("https://github.com/lovecatisgood-sudo");
    expect(source.default).toContain("https://djai.academy");
    expect(source.default).toContain("Siamese Cat Dev</a> from <a");
    expect(source.default).toContain("DJAI Academy</a> &amp; With <a");
    expect(source.default).toContain("DJAI Community</a>");
    expect(source.default).toContain("Free-Opensource-SEO-Screaming-Toad-not-Frog-tool-with-100million-url-crawl-potential");
    expect(source.default).toContain("Support this project on GitHub ★");
  });

  it("keeps raw-only audit summaries compatible with an absent rendering distribution", async () => {
    const source = await import("./App.tsx?raw");
    expect(source.default).toContain("summary.rendering_by_status ?? {}");
    expect(source.default).toContain("summary.rendering_by_status?.completed ?? 0");
  });
});
