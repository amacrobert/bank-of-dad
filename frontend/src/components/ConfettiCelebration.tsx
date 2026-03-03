import { useEffect, useState } from "react";

interface ConfettiCelebrationProps {
  show: boolean;
  onComplete?: () => void;
}

interface Particle {
  id: number;
  x: number;
  y: number;
  color: string;
  size: number;
  rotation: number;
  dx: number;
  dy: number;
}

const COLORS = [
  "var(--color-forest, #2d5a3d)",
  "var(--color-amber, #d4a843)",
  "var(--color-terracotta, #c45d3e)",
  "var(--color-sage, #8fbc8f)",
];

function createParticles(count: number): Particle[] {
  return Array.from({ length: count }, (_, i) => ({
    id: i,
    x: 50 + (Math.random() - 0.5) * 20,
    y: 50 + (Math.random() - 0.5) * 10,
    color: COLORS[Math.floor(Math.random() * COLORS.length)],
    size: 4 + Math.random() * 6,
    rotation: Math.random() * 360,
    dx: (Math.random() - 0.5) * 100,
    dy: -(30 + Math.random() * 60),
  }));
}

export default function ConfettiCelebration({ show, onComplete }: ConfettiCelebrationProps) {
  const [particles, setParticles] = useState<Particle[]>([]);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (show) {
      setParticles(createParticles(50));
      setVisible(true);
      const timer = setTimeout(() => {
        setVisible(false);
        onComplete?.();
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [show, onComplete]);

  if (!visible) return null;

  return (
    <div className="fixed inset-0 pointer-events-none z-50 overflow-hidden">
      {particles.map((p) => (
        <div
          key={p.id}
          className="absolute animate-confetti-fall"
          style={{
            left: `${p.x}%`,
            top: `${p.y}%`,
            width: p.size,
            height: p.size,
            backgroundColor: p.color,
            borderRadius: Math.random() > 0.5 ? "50%" : "2px",
            transform: `rotate(${p.rotation}deg)`,
            // Use CSS custom properties for random trajectory
            "--dx": `${p.dx}vw`,
            "--dy": `${p.dy}vh`,
          } as React.CSSProperties}
        />
      ))}
    </div>
  );
}
