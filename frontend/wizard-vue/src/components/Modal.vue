<script setup lang="ts">
const visible = defineModel();
</script>

<template>
  <Teleport to="#app">
    <div class="modal-wrapper" v-if="visible" @click="visible = false">
      <div class="modal" @click.stop>
        <slot></slot>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
div.modal-wrapper {
  position: fixed;
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
