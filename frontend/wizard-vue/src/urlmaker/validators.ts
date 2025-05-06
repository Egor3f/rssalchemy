import {presetPrefix} from "@/urlmaker/index.ts";
import type {SpecValue} from "@/urlmaker/specs.ts";

export type validator = (v: SpecValue) => boolean;

export function validateUrl(s: SpecValue): boolean {
  let url;
  try {
    url = new URL(s as string);
    return url.protocol === "http:" || url.protocol === "https:"
  } catch {
    return false;
  }
}

export function validatePreset(s: SpecValue): boolean {
  return (s as string).startsWith(presetPrefix);
}

export function validateSelector(s: SpecValue): boolean {
  try {
    document.createDocumentFragment().querySelector(s as string);
    return true;
  } catch {
    return false;
  }
}

export function validateAttribute(s: SpecValue): boolean {
  return /([^\t\n\f \/>"'=]+)/.test(s as string);
}

export function validateDuration(s: SpecValue): boolean {
  return /^\d+[smh]$/.test(s as string);
}
