<script setup lang="ts">
import { onMounted } from "vue";
import FileList from "@/components/web/FileList.vue";
import WebToolbar from "@/components/web/WebToolbar.vue";
import { useWebFilesStore } from "@/stores/webFiles";

const files = useWebFilesStore();

onMounted(() => void files.load(""));
</script>

<template>
  <div class="web-app">
    <WebToolbar />
    <FileList />
  </div>
</template>

<style scoped>
.web-app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
  min-height: 0;
  overflow: hidden;
  overflow-x: hidden;
  background: var(--color-bg-base);
  color: var(--color-text-primary);
}
</style>
