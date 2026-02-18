import { createContext, useContext, useEffect, useState, useCallback, type ReactNode } from "react";
import { get } from "../api";
import { isLoggedIn } from "../auth";
import { getTheme, THEMES, type ThemeDefinition } from "../themes";
import type { ChildUser } from "../types";

interface ThemeContextValue {
  theme: string;
  setTheme: (slug: string) => void;
}

const ThemeContext = createContext<ThemeContextValue>({
  theme: "sapling",
  setTheme: () => {},
});

function applyTheme(themeDef: ThemeDefinition) {
  const root = document.documentElement.style;
  root.setProperty("--color-forest", themeDef.colors.forest);
  root.setProperty("--color-forest-light", themeDef.colors.forestLight);
  root.setProperty("--color-cream", themeDef.colors.cream);
  root.setProperty("--color-cream-dark", themeDef.colors.creamDark);
  document.body.style.backgroundColor = themeDef.colors.cream;
  document.body.style.backgroundImage = themeDef.backgroundSvg;
  document.body.style.backgroundRepeat = "repeat";
  document.body.style.backgroundSize = "auto";
}

function clearTheme() {
  const defaults = THEMES.sapling.colors;
  const root = document.documentElement.style;
  root.setProperty("--color-forest", defaults.forest);
  root.setProperty("--color-forest-light", defaults.forestLight);
  root.setProperty("--color-cream", defaults.cream);
  root.setProperty("--color-cream-dark", defaults.creamDark);
  document.body.style.backgroundColor = "";
  document.body.style.backgroundImage = "";
  document.body.style.backgroundRepeat = "";
  document.body.style.backgroundSize = "";
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState("sapling");

  const setTheme = useCallback((slug: string) => {
    const def = getTheme(slug);
    setThemeState(def.slug);
  }, []);

  // Fetch child theme on mount
  useEffect(() => {
    if (!isLoggedIn()) return;

    get<{ user_type: string; theme?: string | null }>("/auth/me")
      .then((data) => {
        if (data.user_type === "child") {
          const childData = data as ChildUser;
          const def = getTheme(childData.theme);
          setThemeState(def.slug);
        }
        // Parent users stay on sapling (default)
      })
      .catch(() => {
        // Fall back to sapling on error
      });
  }, []);

  // Apply theme CSS whenever it changes
  useEffect(() => {
    const def = getTheme(theme);
    applyTheme(def);

    return () => {
      clearTheme();
    };
  }, [theme]);

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  return useContext(ThemeContext);
}
