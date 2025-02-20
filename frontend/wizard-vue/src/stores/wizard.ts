import {defineStore} from "pinia";
import {emptySpecs, type Field, fields, type Specs} from "@/urlmaker/specs.ts";
import {computed, reactive, ref} from "vue";

export const useWizardStore = defineStore('wizard', () => {
  const specs = reactive(emptySpecs);
  const formValid = computed(() => {
    return fields.every(field => (
      specs[field.name].length === 0 && !(field as Field).required || field.validate(specs[field.name]).ok
    ));
  });
  function updateSpecs(newSpecs: Specs) {
    Object.assign(specs, newSpecs);
  }
  return {specs, formValid, updateSpecs};
});
