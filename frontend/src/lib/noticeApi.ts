import type { AppNotice } from "@/types/app";

export const noticeApi = {
  async drainNotices() {
    const { AppService } = await import("../../bindings/lanfolder");
    return (await AppService.DrainNotices()) as AppNotice[];
  },
  async listenNotices(handler: (notice: AppNotice) => void) {
    const { Events } = await import("@wailsio/runtime");
    return Events.On("app:notice", (event) => handler(event.data as AppNotice));
  },
};
