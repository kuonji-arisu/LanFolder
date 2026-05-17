import { translate } from "@/lib/i18n";

export function userAgentLabel(userAgent: string) {
  if (!userAgent) return translate("common.unknownBrowser");

  const browser = matchBrowser(userAgent);
  const platform = matchPlatform(userAgent);
  return [browser, platform].filter(Boolean).join(" · ") || translate("common.unknownBrowser");
}

function matchBrowser(userAgent: string) {
  if (userAgent.includes("Edg/")) return "Edge";
  if (userAgent.includes("OPR/") || userAgent.includes("Opera/")) return "Opera";
  if (userAgent.includes("Firefox/")) return "Firefox";
  if (userAgent.includes("Chrome/") || userAgent.includes("CriOS/")) return "Chrome";
  if (userAgent.includes("Safari/")) return "Safari";
  return "";
}

function matchPlatform(userAgent: string) {
  if (userAgent.includes("iPhone")) return "iPhone";
  if (userAgent.includes("iPad")) return "iPad";
  if (userAgent.includes("Android")) return "Android";
  if (userAgent.includes("Windows")) return "Windows";
  if (userAgent.includes("Mac OS X")) return "macOS";
  if (userAgent.includes("Linux")) return "Linux";
  return "";
}
