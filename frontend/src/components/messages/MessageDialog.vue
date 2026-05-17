<script setup lang="ts">
import { ref } from "vue";
import { Loader2, RefreshCw, Trash2 } from "lucide-vue-next";
import MessagePanel from "@/components/messages/MessagePanel.vue";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useNoticeStore } from "@/stores/notices";
import { useWebMessagesStore } from "@/stores/webMessages";

const open = defineModel<boolean>("open", { required: true });
const messages = useWebMessagesStore();
const notices = useNoticeStore();
const clearConfirmOpen = ref(false);

async function refreshMessages() {
  notices.showTaskResult(await messages.load());
}

async function sendMessage() {
  notices.showTaskResult(await messages.send());
}

async function clearMessages() {
  const result = await messages.clear();
  notices.showTaskResult(result);
  if (result.ok) {
    clearConfirmOpen.value = false;
    notices.showSuccess("消息已清空");
  }
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
        <div class="message-dialog-actions">
          <Button variant="secondary" size="icon" :disabled="messages.loading" aria-label="刷新消息" @click="refreshMessages">
            <Loader2 v-if="messages.loading" class="h-4 w-4 animate-spin" />
            <RefreshCw v-else class="h-4 w-4" />
          </Button>
          <Button variant="secondary" size="icon" :disabled="messages.loading || !messages.messages.length" aria-label="清空消息" @click="clearConfirmOpen = true">
            <Trash2 class="h-4 w-4" />
          </Button>
        </div>
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

  <Dialog v-model:open="clearConfirmOpen">
    <DialogContent size="sm">
      <DialogHeader>
        <DialogTitle>清空消息</DialogTitle>
        <DialogDescription>这会删除当前共享目录里的传递字符记录。</DialogDescription>
      </DialogHeader>
      <div class="message-clear-actions">
        <Button variant="secondary" :disabled="messages.loading" @click="clearConfirmOpen = false">取消</Button>
        <Button :disabled="messages.loading" @click="clearMessages">
          <Loader2 v-if="messages.loading" class="h-4 w-4 animate-spin" />
          清空
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.message-dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding-right: 44px;
}

.message-dialog-actions,
.message-clear-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.message-clear-actions {
  justify-content: flex-end;
}
</style>
