<script setup lang="ts">
import SpecsForm from "@/components/SpecsForm.vue";
import {ref, watch} from "vue";
import Btn from "@/components/Btn.vue";
import Copyable from "@/components/Copyable.vue";
import EditUrlModal from "@/components/EditUrlModal.vue";
import {decodeUrl, encodeUrl, getScreenshotUrl} from "@/urlmaker";
import {useWizardStore} from "@/stores/wizard.ts";
import {debounce} from "es-toolkit";

const store = useWizardStore();
const existingLink = ref("");
const link = ref("");
const editModalVisible = ref(false);

watch(existingLink, async (value) => {
  if(!value) return;
  existingLink.value = "";
  try {
    store.updateSpecs(await decodeUrl(value));
    link.value = "";
  } catch (e) {
    console.log(e);
    alert(`Decoding error: ${e}`);
  }
});

watch(store.specs, debounce(() => {
    if (store.formValid) {
      generateLink();
    } else {
      link.value = "";
    }
  }, 100),
  {immediate: true}
);

async function generateLink() {
  try {
    link.value = await encodeUrl(store.specs);
  } catch (e) {
    console.log(e);
    alert(`Encoding error: ${e}`);
  }
}

function screenshot() {
  window.open(getScreenshotUrl(store.specs.url));
}

</script>

<template>
  <div class="wrapper">
    <SpecsForm class="specs-form"></SpecsForm>
<!--    <Btn :active="store.formValid" @click="generateLink">Generate link</Btn>-->
    <Btn :active="store.formValid" @click="screenshot">Screenshot</Btn>
    <Btn @click="editModalVisible = true">Edit existing task</Btn>
    <Btn @click="store.reset">Reset Form</Btn>
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
