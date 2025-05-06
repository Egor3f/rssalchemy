<script setup lang="ts">
import {getCurrentInstance} from "vue";

import type {Enum} from "@/common/enum.ts";

const {name, label, values} = defineProps<{
  name: string
  label: string,
  values: Enum,
}>();

const componentId = 'field' + getCurrentInstance()?.uid;

const model = defineModel();

</script>

<template>
  <div class="field">
    <span class="field-label"><label>{{ label }}</label></span>
    <template class="value" v-for="enumValue in values">
      <input
        type="radio"
        :name="name"
        :value="enumValue.value"
        :id="`${componentId}_${enumValue.value}`"
        v-model="model"
      />
      <label class="radio-label" :for="`${componentId}_${enumValue.value}`">{{ enumValue.label }}</label>
    </template>
  </div>
</template>

<style scoped lang="scss">
div.field {
  margin: 0 0 8px 0;
}
.field-label {
  font-size: 0.9em;
  margin-right: 8px;
}
.radio-label {
  font-size: 0.9em;
}
input, .radio-label, .field-label {
  vertical-align: middle;
}
input {
  margin-top: 0;
}
</style>
