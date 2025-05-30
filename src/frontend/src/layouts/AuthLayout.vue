<template>
  <div class="auth-layout">
    <router-view />
  </div>
</template>

<script setup>
// Basic script setup
</script>

<style scoped>
.auth-layout {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f0f2f5; /* Optional background color */
}
</style>