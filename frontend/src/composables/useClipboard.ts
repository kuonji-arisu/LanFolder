import { ref } from "vue";

export function useClipboard(timeout = 1200) {
  const copied = ref(false);

  async function copy(text: string) {
    await navigator.clipboard.writeText(text);
    copied.value = true;
    window.setTimeout(() => {
      copied.value = false;
    }, timeout);
  }

  return { copied, copy };
}
