import { createMemoryHistory, createRouter } from "vue-router";

export const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    {
      path: "/",
      redirect: "/share",
    },
    {
      path: "/share",
      name: "share",
      component: () => import("@/views/ShareView.vue"),
      meta: { title: "LanFolder" },
    },
    {
      path: "/settings",
      name: "settings",
      component: () => import("@/views/SettingsView.vue"),
      meta: { titleKey: "app.settings" },
    },
  ],
});
