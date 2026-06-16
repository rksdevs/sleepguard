import type { Settings } from "./types";

const STORAGE_KEY = "sleepguard.settings";

export function loadSettings(): Settings | null {
  const raw = sessionStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as Settings;
  } catch {
    return null;
  }
}

export function saveSettings(settings: Settings): void {
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
}

export function clearSettings(): void {
  sessionStorage.removeItem(STORAGE_KEY);
}

export function defaultApiBase(): string {
  if (typeof window !== "undefined" && window.location.origin) {
    return window.location.origin;
  }
  return "";
}
