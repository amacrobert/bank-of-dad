const AVATARS = [
  "\u{1F33B}", "\u{1F33F}", "\u{1F342}", "\u{1F338}",
  "\u{1F30A}", "\u{1F319}", "\u2B50",  "\u{1F98B}",
  "\u{1F41D}", "\u{1F344}", "\u{1F438}", "\u{1F98A}",
  "\u{1F43B}", "\u{1F430}", "\u{1F422}", "\u{1F3A8}",
];

interface AvatarPickerProps {
  selected: string | null;
  onSelect: (avatar: string | null) => void;
}

export default function AvatarPicker({ selected, onSelect }: AvatarPickerProps) {
  return (
    <div>
      <label className="block text-sm font-semibold text-bark mb-1.5">Avatar (optional)</label>
      <div className="grid grid-cols-8 gap-1.5">
        {AVATARS.map((emoji) => {
          const isSelected = selected === emoji;
          return (
            <button
              key={emoji}
              type="button"
              onClick={() => onSelect(isSelected ? null : emoji)}
              className={`
                w-9 h-9 flex items-center justify-center text-lg rounded-lg
                transition-all duration-150 cursor-pointer
                ${isSelected
                  ? "bg-forest/15 ring-2 ring-forest scale-110"
                  : "bg-cream hover:bg-cream-dark"
                }
              `}
              aria-label={isSelected ? `Deselect avatar ${emoji}` : `Select avatar ${emoji}`}
              aria-pressed={isSelected}
            >
              {emoji}
            </button>
          );
        })}
      </div>
    </div>
  );
}
