import type {Specs} from "@/urlmaker/specs.ts";

const apiBase = import.meta.env.VITE_API_BASE || document.location.origin;
const renderEndpoint = '/api/v1/render/';  // trailing slash
const screenshotEndpoint = '/api/v1/screenshot';  // no trailing slash

export async function decodeUrl(url: string): Promise<Specs> {
  const splitUrl = url.split(renderEndpoint);
  if(splitUrl.length !== 2) {
    throw 'Split failed';
  }
  let encodedData = splitUrl[1];
  console.log('Data len=' + encodedData.length);
  const m = encodedData.match(/(\d*):?([A-Za-z0-9+/=]+)/);
  if(!m) {
    throw 'Regex failed';
  }
  const version = m[1] ? parseInt(m[1]) : 0;
  console.log('Decoding url using version: ' + version);
  encodedData = m[2];

  let buf = b64decode(encodedData);
  if (version === 0) {
    const jsonData = await decompress(buf);
    return JSON.parse(jsonData);
  }
  throw 'Unknown version'
}

export async function encodeUrl(specs: Specs): Promise<string> {
  const jsonData = JSON.stringify(specs);
  const buf = await compress(jsonData);
  const encodedData = b64encode(buf);
  console.log('Data len=' + encodedData.length);
  const version = 0;
  return `${apiBase}${renderEndpoint}${version}:${encodedData}`
}

export function getScreenshotUrl(url: string): string {
  return `${apiBase}${screenshotEndpoint}?url=${encodeURIComponent(url)}`;
}

function b64encode(buf: Uint8Array): string {
  // @ts-ignore
  const b64str = btoa(String.fromCharCode.apply(null, buf));
  // @ts-ignore
  return b64str.replaceAll('=', '');
}

function b64decode(s: string): Uint8Array {
  return Uint8Array.from(atob(s), c => c.charCodeAt(0));
}

async function compress(s: string): Promise<Uint8Array> {
  let byteArray = new TextEncoder().encode(s);
  let cs = new CompressionStream('deflate-raw');
  let writer = cs.writable.getWriter();
  writer.write(byteArray);
  writer.close();
  let response = new Response(cs.readable);
  return new Uint8Array(await response.arrayBuffer());
}

async function decompress(buf: Uint8Array): Promise<string> {
  let ds = new DecompressionStream('deflate-raw');
  let writer = ds.writable.getWriter();
  writer.write(buf);
  writer.close();
  let response = new Response(ds.readable);
  return response.text();
}
