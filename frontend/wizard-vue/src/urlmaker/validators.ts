import {presetPrefix} from "@/urlmaker/index.ts";

type validResult = { ok: boolean, error?: string };
export type validator = (v: string) => validResult

export function validateUrl(s: string): validResult {
  let url;
  try {
    url = new URL(s);
    return {
      ok: url.protocol === "http:" || url.protocol === "https:",
      error: 'Invalid URL protocol',
    };
  } catch {
    return {ok: false, error: 'Invalid URL'};
  }
}

export function validatePreset(s: string): validResult {
  if(!s.startsWith(presetPrefix)) {
    return {
      ok: false,
      error: 'Not a preset'
    }
  }
  return {ok: true}
}

export function validateOr(...validators: validator[]): validator {
  return function(s: string): validResult {
    return validators.reduce<validResult>((res, v) => {
      let r = v(s);
      if(r.ok) res.ok = true;
      else res.error += r.error + '; ';
      return res;
    }, {ok: false, error: ''});
  }
}

export function validateSelector(s: string): validResult {
  try {
    document.createDocumentFragment().querySelector(s);
    return {ok: true}
  } catch {
    return {ok: false, error: 'Invalid selector'};
  }
}

export function validateDuration(s: string): validResult {
  return {
    ok: /^\d+[smh]$/.test(s),
    error: 'Duration must be number and unit (s/m/h), example: 5s = 5 seconds'
  }
}
