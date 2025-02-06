<script setup lang="ts">
import {fields, type Specs} from '@/urlmaker/specs.ts';
import Field from "@/components/Field.vue";

const model = defineModel<Specs>({required: true});

</script>

<template>
  <div>
    <Field v-for="field in fields" :field="field" v-model="model[field.name]"></Field>
  </div>
</template>

<style scoped lang="scss">

</style>
