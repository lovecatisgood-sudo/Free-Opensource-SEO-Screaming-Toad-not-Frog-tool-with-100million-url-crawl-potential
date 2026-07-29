import { describe, expect, it } from "vitest";
import { foundationStatus } from "./App";

describe("foundationStatus", () => {
  it("exposes each enforced foundation boundary", () => {
    expect(foundationStatus.map(({ phase }) => phase)).toEqual([
      "Guarded network",
      "Durable storage",
      "MCP",
    ]);
  });
});

