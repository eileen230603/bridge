import React from "react";

export type Theme = "light" | "dark";
const KEY = "ap1.theme";

function stored(): Theme | null {
  try {
    const v = localStorage.getItem(KEY);
    return v === "light" || v === "dark" ? v : null;
  } catch {
    return null;
  }
}
const systemTheme = (): Theme =>
  window.matchMedia?.("(prefers-color-scheme: light)").matches ? "light" : "dark";

export function applyTheme(t: Theme) {
  document.documentElement.dataset.theme = t;
}

// Antes de renderizar: preferencia guardada o, si no hay, la del sistema.
applyTheme(stored() ?? systemTheme());

/** Sigue al sistema hasta que el usuario elige manualmente con el botón. */
export function useTheme() {
  const [theme, setTheme] = React.useState<Theme>(() => stored() ?? systemTheme());

  React.useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  React.useEffect(() => {
    const mq = window.matchMedia?.("(prefers-color-scheme: light)");
    if (!mq) return;
    const onChange = () => {
      if (!stored()) setTheme(mq.matches ? "light" : "dark");
    };
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  const toggle = React.useCallback(() => {
    setTheme((t) => {
      const next: Theme = t === "dark" ? "light" : "dark";
      try { localStorage.setItem(KEY, next); } catch { /* sin almacenamiento */ }
      return next;
    });
  }, []);

  return { theme, toggle };
}
