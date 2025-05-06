<script setup lang="ts">
import {fields, InputType, type SpecField} from '@/urlmaker/specs.ts';
import TextField from "@/components/TextField.vue";
import RadioButtons from "@/components/RadioButtons.vue";
import {useWizardStore} from "@/stores/wizard.ts";

const store = useWizardStore();

const groups: SpecField[][] = [];
for (const field of fields) {
  if(groups.length === 0 || groups[groups.length - 1][0].group != field.group) {
    groups.push([field]);
  } else {
    groups[groups.length - 1].push(field);
  }
}

</script>

<template>
  <div>
    <div class="group" v-for="group in Object.values(groups)">
      <template v-for="field in group">
        <TextField
          v-if="field.input_type === InputType.Url || field.input_type === InputType.Text"
          v-show="!field.show_if || field.show_if(store.specs)"
          :name="field.name"
          :label="field.label"
          :input_type="field.input_type"
          :model-value="store.specs[field.name]"
          @update:model-value="event => store.updateSpec(field.name, event)"
        ></TextField>
        <RadioButtons
          v-if="field.input_type === InputType.Radio"
          v-show="!field.show_if || field.show_if(store.specs)"
          :name="field.name"
          :label="field.label"
          :values="field.enum!"
          :model-value="store.specs[field.name]"
          @update:model-value="event => store.updateSpec(field.name, event)"
        ></RadioButtons>
      </template>
    </div>
  </div>
</template>

<style scoped lang="scss">
div.group {
  margin: 20px 0;
}
</style>
