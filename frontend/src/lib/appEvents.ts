import { Events } from "@wailsio/runtime";

const appStateChangedEvent = "app:state-changed";

export const appEvents = {
  onStateChanged(handler: () => void) {
    return Events.On(appStateChangedEvent, handler);
  },
};
