const headerBytes = 4;
export const maximumFrameBytes = 8 * 1024 * 1024;

export interface RenderRequest {
  readonly protocolVersion: 1;
  readonly requestId: string;
  readonly url: string;
  readonly deadlineMs: number;
  readonly maximumRequests: number;
  readonly maximumBytes: number;
}

export interface RenderResponse {
  readonly protocolVersion: 1;
  readonly requestId: string;
  readonly status: "completed" | "blocked" | "failed";
  readonly html?: string;
  readonly errorCode?: string;
}

export function encodeFrame(value: unknown): Uint8Array {
  const payload = Buffer.from(JSON.stringify(value), "utf8");
  if (payload.byteLength > maximumFrameBytes) {
    throw new RangeError("renderer frame exceeds the maximum size");
  }
  const frame = Buffer.allocUnsafe(headerBytes + payload.byteLength);
  frame.writeUInt32BE(payload.byteLength, 0);
  payload.copy(frame, headerBytes);
  return frame;
}

export class FrameDecoder {
  #buffer = Buffer.alloc(0);

  push(chunk: Uint8Array): unknown[] {
    this.#buffer = Buffer.concat([this.#buffer, chunk]);
    const values: unknown[] = [];
    while (this.#buffer.byteLength >= headerBytes) {
      const length = this.#buffer.readUInt32BE(0);
      if (length > maximumFrameBytes) {
        this.#buffer = Buffer.alloc(0);
        throw new RangeError("renderer frame exceeds the maximum size");
      }
      if (this.#buffer.byteLength < headerBytes + length) {
        break;
      }
      const payload = this.#buffer.subarray(headerBytes, headerBytes + length);
      values.push(JSON.parse(payload.toString("utf8")) as unknown);
      this.#buffer = this.#buffer.subarray(headerBytes + length);
    }
    return values;
  }
}

