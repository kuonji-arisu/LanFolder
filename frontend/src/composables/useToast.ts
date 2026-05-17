import { toast } from "vue-sonner";
import type { NoticeLevel } from "@/types/app";

export function useToast() {
  function show(level: NoticeLevel, message: string) {
    if (level === "error") toast.error(message);
    else if (level === "warning") toast.warning(message);
    else if (level === "success") toast.success(message);
    else toast(message);
  }

  return {
    show,
    info: (message: string) => show("info", message),
    success: (message: string) => show("success", message),
    warning: (message: string) => show("warning", message),
    error: (message: string) => show("error", message),
  };
}
