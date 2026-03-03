interface GoalProgressRingProps {
  percent: number;
  size?: number;
  strokeWidth?: number;
  milestone?: boolean;
}

export default function GoalProgressRing({
  percent,
  size = 64,
  strokeWidth = 6,
  milestone = false,
}: GoalProgressRingProps) {
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const clamped = Math.min(100, Math.max(0, percent));
  const offset = circumference - (clamped / 100) * circumference;

  return (
    <div className="relative inline-flex items-center justify-center" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="transform -rotate-90">
        {/* Background track */}
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="var(--color-sand, #d4c9b0)"
          strokeWidth={strokeWidth}
        />
        {/* Progress arc */}
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="var(--color-forest, #2d5a3d)"
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          className="transition-[stroke-dashoffset] duration-600 ease-out"
          style={milestone ? { filter: "drop-shadow(0 0 4px var(--color-forest, #2d5a3d))" } : undefined}
        />
      </svg>
      <span className="absolute text-xs font-bold text-forest tabular-nums">
        {Math.round(clamped)}%
      </span>
    </div>
  );
}
