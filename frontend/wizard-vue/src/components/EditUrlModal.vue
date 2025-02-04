<script setup lang="ts">

import Field from "@/components/Field.vue";
import {type Field as FieldSpec} from "@/urlmaker/specs";
import {validateUrl} from "@/urlmaker/validators.ts";
import Btn from "@/components/Btn.vue";
import {onMounted, onUnmounted, ref, watch} from "vue";

const field: FieldSpec = {
  name: '',
  input_type: 'url',
  label: 'URL of feed for editing',
  default: '',
  required: true,
  validate: validateUrl,
}

const {visible, modelValue} = defineProps({
  visible: Boolean,
  modelValue: {
    type: String,
    required: true,
  }
});

const emit = defineEmits(['close', 'update:modelValue']);

const url = ref(modelValue);
watch(() => visible, () => {
  url.value = modelValue;
});

const valid = ref(false);
watch(url, (value) => {
  valid.value = field.validate(value).ok;
});

const accept = () => {
  valid.value = field.validate(url.value).ok;
  if (valid.value) {
    emit('update:modelValue', url.value);
    emit('close');
  }
}

const listener = (e: KeyboardEvent) => {
  if (e.code === 'Escape') emit('close');
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
  <Teleport to="#app">
    <div class="modal-wrapper" v-if="visible" @click="$emit('close')">
      <div class="modal" @click.stop>
        <Field :field="field" v-model="url" :focused="true"/>
        <Btn :active="valid" @click="accept">Edit</Btn>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
div.modal-wrapper {
  position: absolute;
  left: 0;
  top: 0;
  width: 100vw;
  height: 100vh;
  display: flex;
  background: rgba(200, 200, 200, 0.5);
  backdrop-filter: blur(2px);
}

div.modal {
  width: 100%;
  max-width: 400px;
  margin: auto auto;
  background: #ffffff;
  padding: 10px;
  border-radius: 6px;
  box-shadow: #a0a0a0 1px 2px 2px;
}
</style>
