<script setup lang="ts">
import {getCurrentInstance, onMounted, useTemplateRef} from "vue";

const {name, label, input_type, focused} = defineProps<{
  name: string
  label: string,
  input_type: 'text' | 'url',
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
    <div class="label"><label :for="id">{{ label }}</label></div>
    <div class="input">
      <input :type="input_type" :name="name" :id="id" v-model="model" ref="field"/>
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
