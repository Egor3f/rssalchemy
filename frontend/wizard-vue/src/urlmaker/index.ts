import type {Specs} from "@/urlmaker/specs.ts";
import {b64decode, b64encode, compress, decompress, decompressString} from "@/urlmaker/utils.ts";
import {rssalchemy as pb} from '@/urlmaker/proto/specs.ts';

const apiBase = import.meta.env.VITE_API_BASE || document.location.origin;
const renderEndpoint = '/api/v1/render/';  // trailing slash
const screenshotEndpoint = '/api/v1/screenshot';  // no trailing slash
export const presetPrefix = 'rssalchemy:';

export async function decodeUrl(url: string): Promise<Specs> {
  const splitUrl = url.split(renderEndpoint);
  if(splitUrl.length !== 2) {
    throw 'Split failed';
  }
  let encodedData = splitUrl[1];
  return decodeSpecsPart(encodedData);
}

export async function decodePreset(preset: string): Promise<Specs> {
  if(!preset.startsWith(presetPrefix)) {
    throw 'Invalid preset';
  }
  let encodedData = preset.substring(presetPrefix.length);
  return decodeSpecsPart(encodedData);
}

export async function decodeSpecsPart(encodedData: string): Promise<Specs> {
  console.log('Decoded data len=' + encodedData.length);
  const m = encodedData.match(/(\d*):?([A-Za-z0-9+/=]+)/);
  if(!m) {
    throw 'Regex failed';
  }
  const version = m[1] ? parseInt(m[1]) : 0;
  console.log('Decoding url using version: ' + version);
  encodedData = m[2];

  let buf = b64decode(encodedData);
  switch (version) {
    case 0:
      const jsonData = await decompressString(buf);
      return JSON.parse(jsonData);
    case 1:
      const data = await decompress(buf);
      //@ts-ignore
      return pb.Specs.deserializeBinary(data).toObject();
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
  const pbSpecs = pb.Specs.fromObject(specs);
  let data = pbSpecs.serializeBinary();
  data = await compress(data);
  const encodedData = b64encode(data);
  console.log('Encoded data len=' + encodedData.length);
  const version = 1;
  return `${version}:${encodedData}`;
}

export function getScreenshotUrl(url: string): string {
  return `${apiBase}${screenshotEndpoint}?url=${encodeURIComponent(url)}`;
}
