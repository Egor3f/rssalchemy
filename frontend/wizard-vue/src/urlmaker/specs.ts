import {
  validateAttribute,
  validateDuration,
  validateSelector,
  validateUrl,
  type validator
} from "@/urlmaker/validators.ts";
import {rssalchemy} from "@/urlmaker/proto/specs.ts";
import type {Enum} from "@/common/enum.ts";

export const defaultSpecs = {
  url: '',
  selector_post: '',
  selector_title: '',
  selector_link: '',
  selector_description: '',
  selector_author: '',
  selector_content: '',
  selector_enclosure: '',
  selector_created: '',
  created_extract_from: rssalchemy.ExtractFrom.InnerText,
  created_attribute_name: '',
  cache_lifetime: '10m'
};

export type SpecValue = string | number;
export type Specs = typeof defaultSpecs;

export enum InputType {
  Url = 'url',
  Text = 'text',
  Radio = 'radio'
}

export interface SpecField {
  name: keyof Specs
  input_type: InputType
  enum?: Enum,
  label: string
  validate: validator
  required?: boolean
  group?: string
  show_if?: (specs: Specs) => boolean
}

export const fields: SpecField[] = [
  {
    name: 'url',
    input_type: InputType.Url,
    label: 'URL of page for converting',
    validate: validateUrl,
    required: true,
  },
  {
    name: 'selector_post',
    input_type: InputType.Text,
    label: 'CSS Selector for post',
    validate: validateSelector,
  },
  {
    name: 'selector_title',
    input_type: InputType.Text,
    label: 'CSS Selector for title',
    validate: validateSelector,
  },
  {
    name: 'selector_link',
    input_type: InputType.Text,
    label: 'CSS Selector for link',
    validate: validateSelector,
  },
  {
    name: 'selector_description',
    input_type: InputType.Text,
    label: 'CSS Selector for description',
    validate: validateSelector,
  },
  {
    name: 'selector_author',
    input_type: InputType.Text,
    label: 'CSS Selector for author',
    validate: validateSelector,
  },

  {
    name: 'selector_created',
    input_type: InputType.Text,
    label: 'CSS Selector for created date',
    validate: validateSelector,
    group: 'created',
  },
  {
    name: 'created_extract_from',
    input_type: InputType.Radio,
    enum: [
      {label: 'Inner Text', value: rssalchemy.ExtractFrom.InnerText},
      {label: 'Attribute', value: rssalchemy.ExtractFrom.Attribute},
    ],
    label: 'Extract from',
    validate: value => Object.values(rssalchemy.ExtractFrom).includes(value),
    group: 'created',
    show_if: specs => !!specs.selector_created,
  },
  {
    name: 'created_attribute_name',
    input_type: InputType.Text,
    label: 'Attribute name',
    validate: validateAttribute,
    show_if: specs =>
      !!specs.selector_created && specs.created_extract_from === rssalchemy.ExtractFrom.Attribute,
    group: 'created',
  },

  {
    name: 'selector_content',
    input_type: InputType.Text,
    label: 'CSS Selector for content',
    validate: validateSelector,
  },
  {
    name: 'selector_enclosure',
    input_type: InputType.Text,
    label: 'CSS Selector for enclosure (e.g. image url)',
    validate: validateSelector,
  },
  {
    name: 'cache_lifetime',
    input_type: InputType.Text,
    label: 'Cache lifetime (format examples: 10s, 1m, 2h)',
    validate: validateDuration,
  },
];
