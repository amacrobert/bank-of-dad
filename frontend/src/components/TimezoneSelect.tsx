import { useState, useMemo } from "react";

interface TimezoneOption {
  value: string;
  label: string;
}

const TIMEZONES: TimezoneOption[] = [
  { value: "Pacific/Honolulu", label: "Hawaii (Pacific/Honolulu)" },
  { value: "America/Anchorage", label: "Alaska (America/Anchorage)" },
  { value: "America/Los_Angeles", label: "US Pacific (America/Los_Angeles)" },
  { value: "America/Phoenix", label: "US Mountain - Arizona (America/Phoenix)" },
  { value: "America/Denver", label: "US Mountain (America/Denver)" },
  { value: "America/Chicago", label: "US Central (America/Chicago)" },
  { value: "America/New_York", label: "US Eastern (America/New_York)" },
  { value: "America/Puerto_Rico", label: "Atlantic (America/Puerto_Rico)" },
  { value: "America/Halifax", label: "Canada Atlantic (America/Halifax)" },
  { value: "America/St_Johns", label: "Newfoundland (America/St_Johns)" },
  { value: "America/Sao_Paulo", label: "Brazil (America/Sao_Paulo)" },
  { value: "America/Argentina/Buenos_Aires", label: "Argentina (America/Argentina/Buenos_Aires)" },
  { value: "America/Mexico_City", label: "Mexico City (America/Mexico_City)" },
  { value: "America/Bogota", label: "Colombia (America/Bogota)" },
  { value: "America/Lima", label: "Peru (America/Lima)" },
  { value: "America/Toronto", label: "Canada Eastern (America/Toronto)" },
  { value: "America/Vancouver", label: "Canada Pacific (America/Vancouver)" },
  { value: "America/Winnipeg", label: "Canada Central (America/Winnipeg)" },
  { value: "America/Edmonton", label: "Canada Mountain (America/Edmonton)" },
  { value: "Atlantic/Reykjavik", label: "Iceland (Atlantic/Reykjavik)" },
  { value: "Europe/London", label: "UK (Europe/London)" },
  { value: "Europe/Dublin", label: "Ireland (Europe/Dublin)" },
  { value: "Europe/Paris", label: "Central Europe (Europe/Paris)" },
  { value: "Europe/Berlin", label: "Germany (Europe/Berlin)" },
  { value: "Europe/Madrid", label: "Spain (Europe/Madrid)" },
  { value: "Europe/Rome", label: "Italy (Europe/Rome)" },
  { value: "Europe/Amsterdam", label: "Netherlands (Europe/Amsterdam)" },
  { value: "Europe/Zurich", label: "Switzerland (Europe/Zurich)" },
  { value: "Europe/Stockholm", label: "Sweden (Europe/Stockholm)" },
  { value: "Europe/Helsinki", label: "Finland (Europe/Helsinki)" },
  { value: "Europe/Athens", label: "Greece (Europe/Athens)" },
  { value: "Europe/Istanbul", label: "Turkey (Europe/Istanbul)" },
  { value: "Europe/Moscow", label: "Russia - Moscow (Europe/Moscow)" },
  { value: "Africa/Cairo", label: "Egypt (Africa/Cairo)" },
  { value: "Africa/Johannesburg", label: "South Africa (Africa/Johannesburg)" },
  { value: "Africa/Lagos", label: "Nigeria (Africa/Lagos)" },
  { value: "Asia/Dubai", label: "UAE (Asia/Dubai)" },
  { value: "Asia/Kolkata", label: "India (Asia/Kolkata)" },
  { value: "Asia/Bangkok", label: "Thailand (Asia/Bangkok)" },
  { value: "Asia/Singapore", label: "Singapore (Asia/Singapore)" },
  { value: "Asia/Hong_Kong", label: "Hong Kong (Asia/Hong_Kong)" },
  { value: "Asia/Shanghai", label: "China (Asia/Shanghai)" },
  { value: "Asia/Tokyo", label: "Japan (Asia/Tokyo)" },
  { value: "Asia/Seoul", label: "South Korea (Asia/Seoul)" },
  { value: "Australia/Perth", label: "Australia Western (Australia/Perth)" },
  { value: "Australia/Adelaide", label: "Australia Central (Australia/Adelaide)" },
  { value: "Australia/Sydney", label: "Australia Eastern (Australia/Sydney)" },
  { value: "Pacific/Auckland", label: "New Zealand (Pacific/Auckland)" },
];

interface TimezoneSelectProps {
  value: string;
  onChange: (timezone: string) => void;
}

export default function TimezoneSelect({ value, onChange }: TimezoneSelectProps) {
  const [search, setSearch] = useState("");

  const filtered = useMemo(() => {
    if (!search.trim()) return TIMEZONES;
    const q = search.toLowerCase();
    return TIMEZONES.filter(
      (tz) =>
        tz.label.toLowerCase().includes(q) ||
        tz.value.toLowerCase().includes(q)
    );
  }, [search]);

  return (
    <div className="space-y-1.5">
      <label
        htmlFor="timezone-search"
        className="block text-sm font-semibold text-bark-light"
      >
        Family timezone
      </label>
      <input
        id="timezone-search"
        type="text"
        placeholder="Search timezones..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="
          w-full min-h-[48px] px-4 py-3
          rounded-xl border border-sand bg-white
          text-bark text-base
          transition-all duration-200
          focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest
          placeholder:text-bark-light/50
        "
      />
      <select
        id="timezone-select"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        size={8}
        className="
          w-full px-1 py-1
          rounded-xl border border-sand bg-white
          text-bark text-sm
          transition-all duration-200
          focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest
        "
      >
        {filtered.map((tz) => (
          <option
            key={tz.value}
            value={tz.value}
            className="px-3 py-2 cursor-pointer"
          >
            {tz.label}
          </option>
        ))}
        {filtered.length === 0 && (
          <option disabled value="">
            No timezones match your search
          </option>
        )}
      </select>
      <p className="text-xs text-bark-light/70">
        Currently selected: <span className="font-medium">{TIMEZONES.find((tz) => tz.value === value)?.label || value}</span>
      </p>
    </div>
  );
}
