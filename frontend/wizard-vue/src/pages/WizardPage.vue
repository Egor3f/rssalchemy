<script setup lang="ts">
import SpecsForm from "@/components/SpecsForm.vue";
import {ref, watch} from "vue";
import Btn from "@/components/Btn.vue";
import Copyable from "@/components/Copyable.vue";
import EditUrlModal from "@/components/EditUrlModal.vue";
import {decodePreset, decodeUrl, encodePreset, encodeUrl, getScreenshotUrl} from "@/urlmaker";
import {useWizardStore} from "@/stores/wizard.ts";
import {debounce} from "es-toolkit";
import {validatePreset, validateUrl} from "@/urlmaker/validators.ts";

const store = useWizardStore();
const existingLink = ref("");
const resultLink = ref("");
const resultPreset = ref("");
const editModalVisible = ref(false);

watch(existingLink, async (value) => {
  if(!value) return;
  existingLink.value = "";
  try {
    if(validateUrl(value).ok) store.updateSpecs(await decodeUrl(value));
    else if (validatePreset(value).ok) store.updateSpecs(await decodePreset(value));
  } catch (e) {
    console.log(e);
    alert(`Decoding error: ${e}`);
  }
});

watch(store.specs, debounce(() => {
    if (store.formValid) {
      generate();
    } else {
      resultLink.value = "";
      resultPreset.value = "";
    }
  }, 100),
  {immediate: true}
);

async function generate() {
  try {
    resultLink.value = await encodeUrl(store.specs);
    resultPreset.value = await encodePreset(store.specs);
  } catch (e) {
    console.log(e);
    alert(`Encoding error: ${e}`);
  }
}

function screenshot() {
  if(store.formValid) {
    window.open(getScreenshotUrl(store.specs.url));
  }
}

</script>

<template>
  <div class="wrapper">
    <SpecsForm class="specs-form"></SpecsForm>
<!--    <Btn :active="store.formValid" @click="generateLink">Generate link</Btn>-->
    <Btn :active="store.formValid" @click="screenshot">Screenshot</Btn>
    <Btn @click="editModalVisible = true">Edit existing task / import preset</Btn>
    <Btn @click="store.reset">Reset Form</Btn>
    <div v-if="resultLink" class="link-label">Link for RSS reader:</div>
    <Copyable v-if="resultLink" :contents="resultLink" class="link-view"></Copyable>
    <div v-if="resultPreset" class="link-label">Preset for sharing:</div>
    <Copyable v-if="resultPreset" :contents="resultPreset" class="link-view"></Copyable>
    <EditUrlModal v-model:visible="editModalVisible" v-model="existingLink"></EditUrlModal>
  </div>
</template>

<style scoped lang="scss">
div.wrapper {
  width: 100%;
  max-width: 600px;
  margin: auto;
  padding-bottom: 50px;
}

.specs-form {
  margin-bottom: 15px;
}

.link-label {
  margin-top: 15px;
}
</style>
