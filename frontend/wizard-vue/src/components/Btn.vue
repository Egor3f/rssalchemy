<script setup lang="ts">
  const {active} = defineProps({
    active: Boolean
  });
</script>

<template>
  <div class="wrapper" :class="{active: active}">
    <slot></slot>
  </div>
</template>

<style scoped lang="scss">
div.wrapper {
  display: inline-block;
  border-radius: 4px;
  margin: 4px 8px 0 0;
  padding: 4px 6px;
  user-select: none;

  background-color: #ffffff;
  border: 1px solid #868686;
  color: #505050;

  &.active {
    background-color: #f2f2f2;
    box-shadow: #acacac 1px 1px;
    border: 1px solid #464646;
    color: black;
    cursor: pointer;

    &:hover {
      background-color: #d5d5d5;
    }

    &:active {
      background-color: #c0c0c0;
    }
  }
}
</style>
