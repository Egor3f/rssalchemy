export async function compress(s: string|Uint8Array): Promise<Uint8Array> {
  if(typeof s === 'string') {
    s = new TextEncoder().encode(s);
  }
  let cs = new CompressionStream('deflate-raw');
  let writer = cs.writable.getWriter();
  // noinspection ES6MissingAwait
  writer.write(s);
  // noinspection ES6MissingAwait
  writer.close();
  let response = new Response(cs.readable);
  return new Uint8Array(await response.arrayBuffer());
}

export async function decompress(buf: Uint8Array): Promise<Uint8Array> {
  let ds = new DecompressionStream('deflate-raw');
  let writer = ds.writable.getWriter();
  // noinspection ES6MissingAwait
  writer.write(buf);
  // noinspection ES6MissingAwait
  writer.close();
  let response = new Response(ds.readable);
  return response.bytes();
}

export async function decompressString(buf: Uint8Array): Promise<string> {
  const binary = await decompress(buf);
  // @ts-ignore
  return String.fromCharCode.apply(null, binary);
}

export function b64encode(buf: Uint8Array): string {
  // @ts-ignore
  const b64str = btoa(String.fromCharCode.apply(null, buf));
  // @ts-ignore
  return b64str.replaceAll('=', '');
}

export function b64decode(s: string): Uint8Array {
  return Uint8Array.from(atob(s), c => c.charCodeAt(0));
}
