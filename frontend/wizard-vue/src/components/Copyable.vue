<script setup lang="ts">
import { MdContentCopy } from '@kalimahapps/vue-icons';
import {ref} from "vue";

const {contents} = defineProps({
  contents: String,
});
const copiedTooltip = ref(false);

async function copy() {
  if(contents) {
    await navigator.clipboard.writeText(contents);
    copiedTooltip.value = true;
    setTimeout(() => {
      copiedTooltip.value = false;
    }, 1000);
  }
}

</script>

<template>
  <div class="copyable">
    <span class="contents">{{ contents }}</span>
    <span class="copy" v-if="copiedTooltip">Copied!</span>
    <span class="copy" @click="copy" v-else><MdContentCopy class="icon"/></span>
  </div>
</template>

<style scoped lang="scss">
div.copyable {
  display: flex;
  flex-flow: row nowrap;
  align-items: center;

  border: 1px solid #464646;
  border-radius: 2px;
  margin: 4px 4px 0 0;
  user-select: all;

  span.contents {
    flex: 1;
    padding: 4px;
    overflow: hidden;
    text-align: left;
  }

  span.copy {
    flex: 0;
    cursor: pointer;
    user-select: none;
    padding: 4px;

    .icon {
      display: block;
      font-size: 18px;
    }
  }
}
</style>
