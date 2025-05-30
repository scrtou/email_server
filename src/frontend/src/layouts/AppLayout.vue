<template>
  <!-- This component now acts as a simple pass-through for routes -->
  <!-- The main layout (sidebar, header) is handled by App.vue -->
  <router-view />
</template>

<script setup>
// No script needed if it's just a pass-through
</script>

<style scoped>
/* Styles can be removed or kept minimal if this component has no visual structure itself */
/* For example, if some global padding for routed content was here, it might still be relevant,
   but App.vue's .main-content already has padding. */
</style>