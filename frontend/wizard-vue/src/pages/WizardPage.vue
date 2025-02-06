<script setup lang="ts">
import SpecsForm from "@/components/SpecsForm.vue";
import {reactive, ref, watch} from "vue";
import {type Field, fields, type Specs} from "@/urlmaker/specs.ts";
import Btn from "@/components/Btn.vue";
import Copyable from "@/components/Copyable.vue";
import EditUrlModal from "@/components/EditUrlModal.vue";
import {decodeUrl, encodeUrl, getScreenshotUrl} from "@/urlmaker";

const emptySpecs = fields.reduce((o, f) => {
  o[f.name] = f.default;
  return o
}, {} as Specs);
const specs = reactive(emptySpecs);
const formValid = ref(false);

watch(specs, (value) => {
  formValid.value = fields.every(field => (
    value[field.name].length === 0 && !(field as Field).required || field.validate(value[field.name]).ok
  ));
});

const existingLink = ref("");
const link = ref("");
const editModalVisible = ref(false);

watch(existingLink, async (value) => {
  if(!value) return;
  existingLink.value = "";
  try {
    Object.assign(specs, await decodeUrl(value));
    link.value = "";
  } catch (e) {
    console.log(e);
    alert(`Decoding error: ${e}`);
  }
});

async function generateLink() {
  try {
    link.value = await encodeUrl(specs);
  } catch (e) {
    console.log(e);
    alert(`Encoding error: ${e}`);
  }
}

function screenshot() {
  window.open(getScreenshotUrl(specs.url));
}

</script>

<template>
  <div class="wrapper">
    <SpecsForm v-model="specs" class="specs-form"></SpecsForm>
    <Btn :active="formValid" @click="generateLink">Generate link</Btn>
    <Btn :active="formValid" @click="screenshot">Screenshot</Btn>
    <Btn @click="editModalVisible = true">Edit existing task</Btn>
    <Copyable v-if="link" :contents="link" class="link-view"></Copyable>
    <EditUrlModal :visible="editModalVisible" @close="editModalVisible = false"
                  v-model="existingLink"></EditUrlModal>
  </div>
</template>

<style scoped lang="scss">
div.wrapper {
  width: 100%;
  max-width: 600px;
  margin: auto;
}

.specs-form {
  margin-bottom: 15px;
}

.link-view {
  margin-top: 15px !important;
}
</style>
