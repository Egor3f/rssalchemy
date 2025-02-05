<script setup lang="ts">
import type {Field} from "@/urlmaker/specs.ts";
import {getCurrentInstance, onMounted, useTemplateRef} from "vue";

const {field, focused} = defineProps<{
  field: Field,
  focused?: boolean,
}>();
const id = 'field' + getCurrentInstance()?.uid;
const model = defineModel();

const inputRef = useTemplateRef('field');
onMounted(() => {
  if(focused) inputRef.value?.focus();
})

</script>

<template>
  <div class="field">
    <div class="label"><label :for="id">{{ field.label }}</label></div>
    <div class="input">
      <input :type="field.input_type" :name="field.name" :id="id" v-model="model" ref="field"/>
    </div>
  </div>
</template>

<style scoped lang="scss">
div.field {
  margin: 0 0 8px 0;
}
div.label {
  font-size: 0.9em;
}
div.input {
  margin: 2px 0 0 0;
  box-sizing: border-box;

  input {
    box-sizing: border-box;
    width: 100%;
    padding: 2px;
  }
}
</style>
