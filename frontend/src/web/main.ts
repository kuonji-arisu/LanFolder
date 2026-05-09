import { createApp } from "vue";
import { createPinia } from "pinia";
import WebApp from "./WebApp.vue";
import "@/styles.css";

createApp(WebApp).use(createPinia()).mount("#web-app");
