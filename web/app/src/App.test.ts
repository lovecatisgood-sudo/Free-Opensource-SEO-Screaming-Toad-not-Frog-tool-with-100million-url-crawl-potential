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
  });
});
