<script setup lang="ts">
import SpecsForm from "@/components/SpecsForm.vue";
import {reactive, ref, watch} from "vue";
import {type Field, fields, type Specs} from "@/urlmaker/specs.ts";
import Btn from "@/components/Btn.vue";
import Copyable from "@/components/Copyable.vue";

const emptySpecs = fields.reduce((o, f) => {
  o[f.name] = f.default;
  return o
}, {} as Specs);
const specs = reactive(emptySpecs);
const formValid = ref(false);

watch(specs, (value, oldValue) => {
  formValid.value = fields.every(field => (
    specs[field.name].length === 0 && !(field as Field).required || field.validate(specs[field.name]).ok
  ));
});

const link = ref("https://kek.com");

</script>

<template>
  <SpecsForm v-model="specs" class="specs-form"></SpecsForm>
  <Btn :active="formValid">Generate link</Btn>
  <Btn :active="formValid">Screenshot</Btn>
  <Copyable v-if="link" :contents="link" class="link-view"></Copyable>
</template>

<style scoped lang="scss">
.specs-form {
  margin-bottom: 15px;
}
.link-view {
  margin-top: 15px !important;
}
</style>
