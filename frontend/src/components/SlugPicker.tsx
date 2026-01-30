import { useState, useEffect, useRef } from "react";
import { get } from "../api";
import { SlugCheck } from "../types";

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
    <div className="slug-picker">
      <label htmlFor="slug-input">Choose your family bank URL</label>
      <div className="slug-input-row">
        <span className="slug-prefix">bankofdad.com/</span>
        <input
          id="slug-input"
          type="text"
          value={slug}
          onChange={(e) => handleInput(e.target.value)}
          placeholder="smith-family"
          maxLength={30}
          disabled={disabled}
        />
      </div>

      {checking && <p className="slug-status">Checking availability...</p>}

      {!checking && check && (
        <div className="slug-feedback">
          {check.available ? (
            <p className="slug-available">
              {check.slug} is available!
              <button onClick={() => selectSlug(check.slug)} disabled={disabled}>
                Use this name
              </button>
            </p>
          ) : (
            <div>
              {!check.valid && <p className="slug-invalid">Invalid format. Use 3-30 lowercase letters, numbers, and hyphens.</p>}
              {check.valid && !check.available && <p className="slug-taken">{check.slug} is already taken.</p>}
              {check.suggestions && check.suggestions.length > 0 && (
                <div className="slug-suggestions">
                  <p>Try one of these:</p>
                  <ul>
                    {check.suggestions.map((s) => (
                      <li key={s}>
                        <button onClick={() => selectSlug(s)} disabled={disabled}>
                          {s}
                        </button>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
