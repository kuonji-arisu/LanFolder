import { ref } from "vue";

type Theme = "light" | "dark";

const theme = ref<Theme>((localStorage.getItem("lanfolder-theme") as Theme | null) ?? "light");

function applyTheme(value: Theme) {
  theme.value = value;
  document.documentElement.classList.toggle("dark", value === "dark");
  document.documentElement.setAttribute("data-theme", value);
  localStorage.setItem("lanfolder-theme", value);
}

export function useTheme() {
  return {
    theme,
    setTheme: applyTheme,
    initTheme: () => applyTheme(theme.value),
  };
}
