<script setup lang="ts">
import {fields, type Specs} from '@/urlmaker/specs.ts';
import Field from "@/components/Field.vue";
import {useWizardStore} from "@/stores/wizard.ts";

const store = useWizardStore();

</script>

<template>
  <div>
    <Field v-for="field in fields"
           :field="field"
           :model-value="store.specs[field.name]"
           @update:model-value="event => store.updateSpec(field.name, event)"
    ></Field>
  </div>
</template>

<style scoped lang="scss">

</style>
