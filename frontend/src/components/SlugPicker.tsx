import { useState, useEffect, useRef } from "react";
import { get } from "../api";
import { SlugCheck } from "../types";
import { CheckCircle, XCircle } from "lucide-react";
import LoadingSpinner from "./ui/LoadingSpinner";

interface SlugPickerProps {
  onSelect: (slug: string) => void;
  disabled?: boolean;
}

export default function SlugPicker({ onSelect, disabled }: SlugPickerProps) {
  const [slug, setSlug] = useState("");
  const [check, setCheck] = useState<SlugCheck | null>(null);
  const [checking, setChecking] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    if (slug.length < 3) {
      setCheck(null);
      return;
    }

    setChecking(true);
    debounceRef.current = setTimeout(() => {
      get<SlugCheck>(`/families/check-slug?slug=${encodeURIComponent(slug)}`)
        .then((result) => {
          setCheck(result);
          setChecking(false);
        })
        .catch(() => {
          setChecking(false);
        });
    }, 300);

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, [slug]);

  const handleInput = (value: string) => {
    const normalized = value.toLowerCase().replace(/[^a-z0-9-]/g, "");
    setSlug(normalized);
  };

  const selectSlug = (s: string) => {
    setSlug(s);
    onSelect(s);
  };

  return (
    <div className="space-y-3">
      <label htmlFor="slug-input" className="block text-sm font-semibold text-bark-light">
        Choose your family bank URL
      </label>
      <div className="flex items-center rounded-xl border border-sand bg-white overflow-hidden focus-within:ring-2 focus-within:ring-forest/30 focus-within:border-forest transition-all">
        <span className="px-3 py-3 bg-cream-dark text-bark-light text-sm font-medium border-r border-sand whitespace-nowrap">
          bankofdad.com/
        </span>
        <input
          id="slug-input"
          type="text"
          value={slug}
          onChange={(e) => handleInput(e.target.value)}
          placeholder="smith-family"
          maxLength={30}
          disabled={disabled}
          className="flex-1 min-h-[48px] px-3 py-3 bg-transparent text-bark text-base placeholder:text-bark-light/50 focus:outline-none disabled:cursor-not-allowed"
        />
      </div>

      {checking && <LoadingSpinner variant="inline" message="Checking availability..." />}

      {!checking && check && (
        <div>
          {check.available ? (
            <div className="flex items-center justify-between gap-2">
              <span className="inline-flex items-center gap-1.5 text-sm text-forest font-medium">
                <CheckCircle className="h-4 w-4" aria-hidden="true" />
                {check.slug} is available!
              </span>
              <button
                onClick={() => selectSlug(check.slug)}
                disabled={disabled}
                className="text-sm font-semibold text-forest hover:text-forest-light transition-colors cursor-pointer disabled:opacity-50"
              >
                Use this name
              </button>
            </div>
          ) : (
            <div className="space-y-2">
              {!check.valid && (
                <span className="inline-flex items-center gap-1.5 text-sm text-terracotta font-medium">
                  <XCircle className="h-4 w-4" aria-hidden="true" />
                  Invalid format. Use 3-30 lowercase letters, numbers, and hyphens.
                </span>
              )}
              {check.valid && !check.available && (
                <span className="inline-flex items-center gap-1.5 text-sm text-terracotta font-medium">
                  <XCircle className="h-4 w-4" aria-hidden="true" />
                  {check.slug} is already taken.
                </span>
              )}
              {check.suggestions && check.suggestions.length > 0 && (
                <div>
                  <p className="text-sm text-bark-light mb-2">Try one of these:</p>
                  <div className="flex flex-wrap gap-2">
                    {check.suggestions.map((s) => (
                      <button
                        key={s}
                        onClick={() => selectSlug(s)}
                        disabled={disabled}
                        className="px-3 py-1.5 rounded-full bg-sage-light/30 text-sm font-medium text-forest hover:bg-sage-light/50 transition-colors cursor-pointer disabled:opacity-50"
                      >
                        {s}
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
