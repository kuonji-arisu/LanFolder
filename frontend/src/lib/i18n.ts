import { computed, ref } from "vue";
import en from "@i18n-data/en.json";
import zhCN from "@i18n-data/zh-CN.json";

export type Language = "zh-CN" | "en";
type Params = Record<string, unknown>;
type Dictionary = Record<string, string>;

const dictionaries: Record<Language, Dictionary> = {
  "zh-CN": zhCN,
  en,
};

const currentLanguage = ref<Language>("zh-CN");

export const languageOptions = [
  { value: "zh-CN" as const, label: "中文" },
  { value: "en" as const, label: "English" },
];

export function normalizeLanguage(language?: string | null): Language {
  const value = String(language ?? "").trim().toLowerCase().replace("_", "-");
  if (value.startsWith("en")) return "en";
  if (value.startsWith("zh")) return "zh-CN";
  return "zh-CN";
}

export function setLanguage(language?: string | null) {
  currentLanguage.value = normalizeLanguage(language);
  document.documentElement.setAttribute("lang", currentLanguage.value);
}

export function translate(key: string, params?: Params) {
  const language = currentLanguage.value;
  const value = dictionaries[language][key] ?? (language !== "en" ? dictionaries.en[key] : undefined) ?? key;
  return interpolate(value, params);
}

export function useI18n() {
  return {
    language: computed(() => currentLanguage.value),
    languageOptions,
    setLanguage,
    t: translate,
  };
}

function interpolate(value: string, params?: Params) {
  if (!params) return value;
  return value.replace(/\{(\w+)\}/g, (match, key) => (key in params ? String(params[key]) : match));
}
