import type { AppNotice, NoticePresentation } from "@/types/app";
import type { Notice as WailsNotice } from "../../bindings/lanfolder/internal/desktop/models";

export const noticeApi = {
  async drainNotices() {
    const { AppService } = await import("../../bindings/lanfolder");
    return (await AppService.DrainNotices()) as AppNotice[];
  },
  async listenNotices(handler: (notice: AppNotice) => void) {
    const { Events } = await import("@wailsio/runtime");
    return Events.On("app:notice", (event) => handler(event.data as AppNotice));
  },
  async presentNotice(notice: AppNotice, message: string): Promise<NoticePresentation> {
    const { AppService } = await import("../../bindings/lanfolder");
    return (await AppService.PresentNotice(notice as unknown as WailsNotice, message)) as NoticePresentation;
  },
};
