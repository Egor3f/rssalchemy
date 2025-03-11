<script setup lang="ts">
import { RouterView } from 'vue-router'
import ForkMe from "@/components/ForkMe.vue";
</script>

<template>
  <div id="header">
    <div id="logo">
      <img src="./assets/logo.png" alt="logo"/>
    </div>
  </div>
  <RouterView />
  <ForkMe></ForkMe>
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
