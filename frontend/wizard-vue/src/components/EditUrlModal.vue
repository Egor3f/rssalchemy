<script setup lang="ts">

import Field from "@/components/Field.vue";
import {type Field as FieldSpec} from "@/urlmaker/specs";
import {validateOr, validatePreset, validateUrl} from "@/urlmaker/validators.ts";
import Btn from "@/components/Btn.vue";
import {onMounted, onUnmounted, ref, watch} from "vue";
import Modal from "@/components/Modal.vue";

const field: FieldSpec = {
  name: '',
  input_type: 'url',
  label: 'URL of feed or preset',
  default: '',
  required: true,
  validate: validateOr(validateUrl, validatePreset),
}

const visible = defineModel('visible', {
  type: Boolean,
  required: true
});
const modelValue = defineModel({
  type: String,
  required: true
});
const emit = defineEmits(['update:modelValue', 'update:visible']);

const url = ref(modelValue.value);
watch(visible, () => {
  url.value = modelValue.value;
});

const valid = ref(false);
watch(url, (value) => {
  valid.value = field.validate(value).ok;
});

const accept = () => {
  valid.value = field.validate(url.value).ok;
  if (valid.value) {
    emit('update:modelValue', url.value);
    emit('update:visible', false);
  }
}

const listener = (e: KeyboardEvent) => {
  if (e.code === 'Escape') emit('update:visible', false);
  if (e.code === 'Enter') accept();
};
onMounted(() => {
  document.addEventListener('keyup', listener);
});
onUnmounted(() => {
  document.removeEventListener('keyup', listener);
});

</script>

<template>
  <Modal v-model="visible">
    <Field :field="field" v-model="url" :focused="true"/>
    <Btn :active="valid" @click="accept">Edit</Btn>
  </Modal>
</template>

<style scoped lang="scss">

</style>
