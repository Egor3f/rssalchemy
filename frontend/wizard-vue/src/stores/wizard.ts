import {defineStore} from "pinia";
import {emptySpecs, type Field, type FieldNames, fields, type Specs} from "@/urlmaker/specs.ts";
import {computed, reactive} from "vue";
import {debounce} from "es-toolkit";

const LOCAL_STORAGE_KEY = 'rssalchemy_store_wizard';

export const useWizardStore = defineStore('wizard', () => {

  const locStorageContent = localStorage.getItem(LOCAL_STORAGE_KEY);
  const initialSpecs = locStorageContent ? JSON.parse(locStorageContent) as Specs : emptySpecs;

  const specs = reactive(Object.assign({}, initialSpecs));

  const formValid = computed(() => {
    return fields.every(field => (
      specs[field.name].length === 0 && !(field as Field).required || field.validate(specs[field.name]).ok
    ));
  });

  const updateLocalStorage = debounce(() => {
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(specs));
  }, 250);

  function updateSpec(fieldName: FieldNames, newValue: string) {
    specs[fieldName] = newValue;
    updateLocalStorage();
  }
  function updateSpecs(newValue: Specs) {
    Object.assign(specs, newValue);
    updateLocalStorage();
  }
  function reset() {
    Object.assign(specs, emptySpecs);
    updateLocalStorage();
  }

  return {specs, formValid, updateSpec, updateSpecs, reset};
});
