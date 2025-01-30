<script setup lang="ts">
import type {Field} from "@/urlmaker/specs.ts";
import {getCurrentInstance} from "vue";

defineProps<{
  field: Field
}>();
const id = 'field' + getCurrentInstance()?.uid;
const model = defineModel();

</script>

<template>
  <div class="field">
    <div class="label"><label :for="id">{{ field.label }}</label></div>
    <div class="input">
      <input :type="field.input_type" :name="field.name" :id="id" v-model="model"/>
    </div>
  </div>
</template>

<style scoped lang="scss">
div.field {
  margin: 0 0 8px 0;
}
</style>
