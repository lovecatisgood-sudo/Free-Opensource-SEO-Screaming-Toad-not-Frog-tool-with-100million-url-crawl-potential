import { describe, expect, it } from "vitest";
import { encodeFrame, FrameDecoder, maximumFrameBytes } from "./protocol.js";

describe("renderer framing", () => {
  it("decodes fragmented and consecutive messages", () => {
    const first = encodeFrame({ requestId: "one" });
    const second = encodeFrame({ requestId: "two" });
    const combined = Buffer.concat([first, second]);
    const decoder = new FrameDecoder();

    expect(decoder.push(combined.subarray(0, 3))).toEqual([]);
    expect(decoder.push(combined.subarray(3))).toEqual([
      { requestId: "one" },
      { requestId: "two" },
    ]);
  });

  it("rejects oversized advertised frames before buffering the body", () => {
    const header = Buffer.alloc(4);
    header.writeUInt32BE(maximumFrameBytes + 1);
    expect(() => new FrameDecoder().push(header)).toThrow(RangeError);
  });
});

