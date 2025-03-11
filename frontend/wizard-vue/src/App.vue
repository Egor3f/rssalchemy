<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div id="header">
    <div id="logo">
      <img src="./assets/logo.png" alt="logo"/>
    </div>
  </div>
  <RouterView />
</template>

<style scoped lang="scss">
#header {
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;

  margin: 5px 0 10px 0;

  #logo {
    img {
      max-width: 300px;
    }
  }
}
</style>
