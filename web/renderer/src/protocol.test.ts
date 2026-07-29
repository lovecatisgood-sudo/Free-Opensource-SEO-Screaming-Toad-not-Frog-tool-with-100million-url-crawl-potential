import { describe, expect, it } from "vitest";
import { encodeFrame, FrameDecoder, maximumFrameBytes, validateRenderRequest } from "./protocol.js";

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

  it("rejects unknown fields and budgets outside the protocol",()=>{const valid={kind:"render_request",protocolVersion:1,requestId:"r1",url:"https://example.com/",deadlineMs:10_000,maximumRequests:100,maximumBytes:1_000_000} as const;expect(validateRenderRequest(valid)).toEqual(valid);expect(()=>validateRenderRequest({...valid,unexpected:true})).toThrow();expect(()=>validateRenderRequest({...valid,maximumRequests:10_000})).toThrow();});
});
