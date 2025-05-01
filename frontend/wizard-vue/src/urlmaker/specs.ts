import {
  validateDuration,
  validateSelector,
  validateUrl,
  type validator
} from "@/urlmaker/validators.ts";

export interface Field {
  name: string
  input_type: string
  label: string
  default: string
  validate: validator
  required?: boolean
}

export const fields = [
  {
    name: 'url',
    input_type: 'url',
    label: 'URL of page for converting',
    default: '',
    validate: validateUrl,
    required: true,
  },
  {
    name: 'selector_post',
    input_type: 'text',
    label: 'CSS Selector for post',
    default: '',
    validate: validateSelector,
  },
  {
    name: 'selector_title',
    input_type: 'text',
    label: 'CSS Selector for title',
    default: '',
    validate: validateSelector,
  },
  {
    name: 'selector_link',
    input_type: 'text',
    label: 'CSS Selector for link',
    default: '',
    validate: validateSelector,
  },
  {
    name: 'selector_description',
    input_type: 'text',
    label: 'CSS Selector for description',
    default: '',
    validate: validateSelector,
  },
  {
    name: 'selector_author',
    input_type: 'text',
    label: 'CSS Selector for author',
    default: '',
    validate: validateSelector,
  },
  {
    name: 'selector_created',
    input_type: 'text',
    label: 'CSS Selector for created date',
    default: '',
    validate: validateSelector,
  },
  {
    name: 'selector_content',
    input_type: 'text',
    label: 'CSS Selector for content',
    default: '',
    validate: validateSelector,
  },
  {
    name: 'selector_enclosure',
    input_type: 'text',
    label: 'CSS Selector for enclosure (e.g. image url)',
    default: '',
    validate: validateSelector,
  },
  {
    name: 'cache_lifetime',
    input_type: 'text',
    label: 'Cache lifetime (format examples: 10s, 1m, 2h)',
    default: '10m',
    validate: validateDuration,
  },
] as const satisfies Field[];

export type FieldNames = (typeof fields)[number]['name'];

export type Specs = {[k in FieldNames]: string};

export const emptySpecs = fields.reduce((o, f) => {
  o[f.name] = f.default;
  return o
}, {} as Specs);
