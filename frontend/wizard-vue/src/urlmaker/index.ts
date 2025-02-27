import type {Specs} from "@/urlmaker/specs.ts";
import crc32 from 'crc/crc32';

const apiBase = import.meta.env.VITE_API_BASE || document.location.origin;
const renderEndpoint = '/api/v1/render/';  // trailing slash
const screenshotEndpoint = '/api/v1/screenshot';  // no trailing slash
export const presetPrefix = 'rssalchemy:';

export async function decodeUrl(url: string): Promise<Specs> {
  const splitUrl = url.split(renderEndpoint);
  if (splitUrl.length !== 2) {
    throw 'Split failed';
  }
  let encodedData = splitUrl[1];
  return decodeSpecsPart(encodedData);
}

export async function decodePreset(preset: string): Promise<Specs> {
  if (!preset.startsWith(presetPrefix)) {
    throw 'Invalid preset';
  }
  let encodedData = preset.substring(presetPrefix.length);
  return decodeSpecsPart(encodedData);
}

export async function decodeSpecsPart(encodedData: string): Promise<Specs> {
  console.log('Data len=' + encodedData.length);
  const m = encodedData.match(/(\d*):?([A-Za-z0-9+/=]+)/);
  if (!m) {
    throw 'Regex failed';
  }
  const version = m[1] ? parseInt(m[1]) : 0;
  console.log('Decoding url using version: ' + version);
  encodedData = m[2];

  let buf = b64decode(encodedData);
  let jsonData;
  switch (version) {
    case 0:
      jsonData = await decompress(buf);
      return JSON.parse(jsonData);
    case 1:
      jsonData = await decompress(buf);
      jsonData = decodeDict(jsonData);
      jsonData = unescapeUnicode(jsonData);
      jsonData = stripCrc(jsonData);
      return JSON.parse(jsonData);
    default:
      throw 'Unknown version'
  }
}

export async function encodeUrl(specs: Specs): Promise<string> {
  return `${apiBase}${renderEndpoint}${await encodeSpecsPart(specs)}`
}

export async function encodePreset(specs: Specs): Promise<string> {
  return `${presetPrefix}${await encodeSpecsPart(specs)}`;
}

export async function encodeSpecsPart(specs: Specs): Promise<string> {
  let data = JSON.stringify(specs);
  console.debug(`Before dict-encoding: ${data.length}`);
  // data = addCrc(data);
  // data = escapeUnicode(data);
  // data = encodeDict(data);
  console.debug(`After dict-encoding: ${data.length}`);
  const buf = await compress(data);
  data = b64encode(buf);
  console.log('Compressed data len=' + data.length);
  const version = 1;
  return `${version}:${data}`;
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
  // noinspection ES6MissingAwait
  writer.write(byteArray);
  // noinspection ES6MissingAwait
  writer.close();
  let response = new Response(cs.readable);
  return new Uint8Array(await response.arrayBuffer());
}

async function decompress(buf: Uint8Array): Promise<string> {
  let ds = new DecompressionStream('deflate-raw');
  let writer = ds.writable.getWriter();
  // noinspection ES6MissingAwait
  writer.write(buf);
  // noinspection ES6MissingAwait
  writer.close();
  let response = new Response(ds.readable);
  return response.text();
}

function escapeUnicode(s: string): string {
  return s
    .split("")
    .map(function (c): string {
      const code = c.charCodeAt(0);
      return code < 128 ? c : "\\u" + code.toString(16).padStart(4, "0");
    })
    .join("");
}

function unescapeUnicode(s: string): string {
  return s.replace(/\\u([\dA-Fa-f]{4})/g, (_, hex) => (
    String.fromCharCode(parseInt(hex, 16))
  ));
}

// Make in sync with go code. If changing dict - increase encoding 'version'.
const dict = [
  // '","',
  // '":"',
  // '"selector_',
  // '"http://',
  // '"https://',
  // '"https://www.',
  // 'div',
  // 'span',
  // 'img',
  // 'feed',
  // 'post',
  // 'title',
  // 'nth-of-type(',
  // 'nth-child(',

  "selector_post", "url", "selector_title", "selector_link", "selector_description", "selector_author", "selector_created", "selector_content", "selector_enclosure", "cache_lifetime"
];

function encodeDict(s: string): string {
  Array.from(s).map(c => {
    if (c.codePointAt(0)! >= 128) {
      throw 'Non-ascii char found; string must be escaped'
    }
  });
  let sorted = Array.from(dict).sort((a, b) => b.length - a.length);
  for (let i = 0; i < sorted.length; i++) {
    s = s.replaceAll(sorted[i], String.fromCodePoint(128 + i));
  }
  return s;
}

function decodeDict(s: string): string {
  return Array.from(s).map(c => {
    let code = c.codePointAt(0)!;
    if(code < 128) return c;
    code -= 128;
    if(code >= dict.length) {
      throw 'char code > dict length';
    }
    return dict[code];
  }).join("");
}

function addCrc(s: string): string {
  return s + crc32(s).toString(16).padStart(8, "0");
}

function stripCrc(s: string): string {
  let sum = s.substring(s.length - 8);
  s = s.substring(0, s.length - 8);
  if(crc32(s) !== parseInt(sum, 16)) {
    throw 'Checksum mismatch';
  }
  return s;
}
