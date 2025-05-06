<script setup lang="ts">
import {fields, InputType} from '@/urlmaker/specs.ts';
import TextField from "@/components/TextField.vue";
import {useWizardStore} from "@/stores/wizard.ts";
import {groupBy} from "es-toolkit";

const store = useWizardStore();

const groups = groupBy(fields, item => item.group || '');

</script>

<template>
  <div>
    <div class="group" v-for="group in Object.values(groups)">
      <template v-for="field in group">
        <TextField
          v-if="field.input_type === InputType.Url || field.input_type === InputType.Text"
          :name="field.name"
          :label="field.label"
          :input_type="field.input_type"
          :model-value="store.specs[field.name]"
          @update:model-value="event => store.updateSpec(field.name, event)"
        ></TextField>
      </template>
    </div>
  </div>
</template>

<style scoped lang="scss">
div.group {
  margin: 6px 0;
}
</style>
