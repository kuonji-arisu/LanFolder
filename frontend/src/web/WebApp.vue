<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import AccessGate from "@/components/web/AccessGate.vue";
import FileList from "@/components/web/FileList.vue";
import MessageDialog from "@/components/messages/MessageDialog.vue";
import WebToolbar from "@/components/web/WebToolbar.vue";
import { Toaster } from "@/components/ui/sonner";
import { useTheme } from "@/composables/useTheme";
import { useI18n } from "@/lib/i18n";
import { useWebFilesStore } from "@/stores/webFiles";
import { useWebMessagesStore } from "@/stores/webMessages";

const files = useWebFilesStore();
const messages = useWebMessagesStore();
const { initTheme } = useTheme();
const { t } = useI18n();
const messagesOpen = ref(false);
const authorized = ref(false);
const accessReady = ref(false);

onMounted(() => {
  initTheme();
});

watch(messagesOpen, (open) => {
  if (open) void messages.load();
});

function handleAuthorized() {
  authorized.value = true;
  accessReady.value = true;
  void files.load("");
}
</script>

<template>
  <div class="web-app">
    <AccessGate v-if="!authorized" @authorized="handleAuthorized" />

    <template v-if="authorized && accessReady">
      <WebToolbar :messages-open="messagesOpen" @toggle-messages="messagesOpen = !messagesOpen" />
      <MessageDialog
        v-model:open="messagesOpen"
        v-model:draft="messages.draft"
        :messages="messages.messages"
        :current-client-id="messages.clientId"
        :loading="messages.loading"
        :can-send="files.canUpload"
        :can-clear="files.canDelete"
        :send-disabled-text="t('message.sendRequiresUpload')"
        :load-messages="messages.load"
        :send-message="messages.send"
        :clear-messages="messages.clear"
      />
      <FileList />
    </template>
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
