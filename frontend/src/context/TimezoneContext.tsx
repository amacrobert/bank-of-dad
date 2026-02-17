import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { getSettings } from "../api";
import { isLoggedIn } from "../auth";

interface TimezoneContextValue {
  timezone: string;
  loading: boolean;
}

const TimezoneContext = createContext<TimezoneContextValue>({
  timezone: "UTC",
  loading: false,
});

export function TimezoneProvider({ children }: { children: ReactNode }) {
  const [timezone, setTimezone] = useState("UTC");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!isLoggedIn()) {
      setLoading(false);
      return;
    }

    getSettings()
      .then((settings) => {
        if (settings.timezone) {
          setTimezone(settings.timezone);
        }
      })
      .catch(() => {
        // Fall back to UTC on error (e.g., child user without settings access)
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  return (
    <TimezoneContext.Provider value={{ timezone, loading }}>
      {children}
    </TimezoneContext.Provider>
  );
}

export function useTimezone(): string {
  return useContext(TimezoneContext).timezone;
}
