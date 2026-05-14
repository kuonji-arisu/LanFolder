<script setup lang="ts">
import { Loader2, RefreshCw } from "lucide-vue-next";
import MessagePanel from "@/components/messages/MessagePanel.vue";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useErrorToast } from "@/composables/useErrorToast";
import { useWebMessagesStore } from "@/stores/webMessages";

const open = defineModel<boolean>("open", { required: true });
const messages = useWebMessagesStore();
const { showResultError } = useErrorToast();

async function refreshMessages() {
  showResultError(await messages.load());
}

async function sendMessage() {
  showResultError(await messages.send());
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent>
      <div class="message-dialog-head">
        <DialogHeader>
          <DialogTitle>传递字符</DialogTitle>
          <DialogDescription>共享文件旁的临时文字通道，手动刷新同步消息</DialogDescription>
        </DialogHeader>
        <Button variant="secondary" size="icon" :disabled="messages.loading" aria-label="刷新消息" @click="refreshMessages">
          <Loader2 v-if="messages.loading" class="h-4 w-4 animate-spin" />
          <RefreshCw v-else class="h-4 w-4" />
        </Button>
      </div>
      <MessagePanel
        v-model:draft="messages.draft"
        :messages="messages.messages"
        :current-client-id="messages.clientId"
        :loading="messages.loading"
        empty-text="还没有传递字符"
        @send="sendMessage"
      />
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.message-dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding-right: 48px;
}
</style>
