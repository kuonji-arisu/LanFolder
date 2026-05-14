<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import FileList from "@/components/web/FileList.vue";
import MessageDialog from "@/components/messages/MessageDialog.vue";
import WebToolbar from "@/components/web/WebToolbar.vue";
import { Toaster } from "@/components/ui/sonner";
import { useWebFilesStore } from "@/stores/webFiles";
import { useWebMessagesStore } from "@/stores/webMessages";

const files = useWebFilesStore();
const messages = useWebMessagesStore();
const messagesOpen = ref(false);

onMounted(() => void files.load(""));

watch(messagesOpen, (open) => {
  if (open && !messages.messages.length) void messages.load();
});
</script>

<template>
  <div class="web-app">
    <WebToolbar :messages-open="messagesOpen" @toggle-messages="messagesOpen = !messagesOpen" />
    <MessageDialog v-model:open="messagesOpen" />
    <FileList />
    <Toaster position="top-right" />
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
