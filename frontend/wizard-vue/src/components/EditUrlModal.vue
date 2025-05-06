<script setup lang="ts">

import TextField from "@/components/inputs/TextField.vue";
import Btn from "@/components/Btn.vue";
import {onMounted, onUnmounted, ref, watch} from "vue";
import Modal from "@/components/Modal.vue";
import {validatePreset, validateUrl} from "@/urlmaker/validators.ts";

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
  valid.value = validateUrl(value) || validatePreset(value);
});

const accept = () => {
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
    <TextField
      name="url"
      input_type="url"
      label="URL of feed or preset"
      v-model="url"
      :focused="true"
    />
    <Btn :active="valid" @click="accept">Edit</Btn>
  </Modal>
</template>

<style scoped lang="scss">

</style>
