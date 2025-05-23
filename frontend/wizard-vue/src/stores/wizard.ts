import {defineStore} from "pinia";
import {defaultSpecs, type SpecField, fields, type Specs, type SpecValue} from "@/urlmaker/specs.ts";
import {computed, reactive} from "vue";
import {debounce} from "es-toolkit";

const LOCAL_STORAGE_KEY = 'rssalchemy_store_wizard';

export const useWizardStore = defineStore('wizard', () => {

  const locStorageContent = localStorage.getItem(LOCAL_STORAGE_KEY);
  const initialSpecs = locStorageContent ? JSON.parse(locStorageContent) as Specs : defaultSpecs;

  const specs = reactive(Object.assign({}, initialSpecs));

  const formValid = computed(() => {
    return fields.every(field => (
      !specs[field.name] && !(field as SpecField).required || field.validate(specs[field.name]!)
    ));
  });

  const updateLocalStorage = debounce(() => {
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(specs));
  }, 100);

  function updateSpec<K extends keyof Specs>(fieldName: K, newValue: Specs[K]) {
    specs[fieldName] = newValue;
    updateLocalStorage();
  }
  function updateSpecs(newValue: Specs) {
    Object.assign(specs, newValue);
    updateLocalStorage();
  }
  function reset() {
    Object.assign(specs, defaultSpecs);
    updateLocalStorage();
  }

  return {specs, formValid, updateSpec, updateSpecs, reset};
});
