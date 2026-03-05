import { useState, useEffect } from "react";

interface BankNameInputProps {
  value: string;
  onChange: (name: string) => void;
}

const QUICK_OPTIONS = ["Dad", "Mom"];
const MAX_LENGTH = 12;
const VALID_PATTERN = /^[\p{L}\p{N} '-]*$/u;

export default function BankNameInput({ value, onChange }: BankNameInputProps) {
  const [inputValue, setInputValue] = useState(value);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setInputValue(value);
  }, [value]);

  const isQuickOption = QUICK_OPTIONS.includes(value);

  const handleQuickSelect = (name: string) => {
    setInputValue(name);
    setError(null);
    onChange(name);
  };

  const handleInputChange = (text: string) => {
    setInputValue(text);

    if (text.length > MAX_LENGTH) {
      setError(`Maximum ${MAX_LENGTH} characters`);
      return;
    }

    if (text && !VALID_PATTERN.test(text)) {
      setError("Only letters, numbers, spaces, hyphens, and apostrophes");
      return;
    }

    setError(null);
    onChange(text);
  };

  return (
    <div className="space-y-4">
      {/* Preview */}
      <div className="text-center">
        <span className="text-2xl font-bold text-forest">
          Bank of {value || "___"}
        </span>
      </div>

      {/* Quick select */}
      <div className="flex justify-center gap-3">
        {QUICK_OPTIONS.map((option) => (
          <button
            key={option}
            type="button"
            onClick={() => handleQuickSelect(option)}
            className={`
              rounded-full px-5 py-2 text-sm font-semibold transition-colors cursor-pointer
              ${value === option
                ? "bg-forest text-white"
                : "bg-white text-bark-light border border-sand hover:bg-cream-dark"
              }
            `}
          >
            {option}
          </button>
        ))}
      </div>

      {/* Custom input */}
      <div className="space-y-1.5">
        <label htmlFor="bank-name-input" className="block text-sm font-semibold text-bark-light">
          Or type your own
        </label>
        <div className="relative">
          <input
            id="bank-name-input"
            type="text"
            value={isQuickOption ? "" : inputValue}
            onChange={(e) => handleInputChange(e.target.value)}
            placeholder="e.g. Auntie, Grandpa, Nana"
            maxLength={MAX_LENGTH}
            className={`
              w-full min-h-[48px] px-4 py-3
              rounded-xl border border-sand bg-white
              text-bark text-base placeholder:text-bark-light/50
              transition-all duration-200
              focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest
              ${error ? "border-terracotta ring-1 ring-terracotta/30" : ""}
            `}
          />
          <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-bark-light/60">
            {(isQuickOption ? 0 : inputValue.length)}/{MAX_LENGTH}
          </span>
        </div>
        {error && (
          <p className="text-sm text-terracotta font-medium">{error}</p>
        )}
      </div>
    </div>
  );
}
