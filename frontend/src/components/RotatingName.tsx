import { useState, useEffect, useRef } from "react";

const NAMES = [
  { name: "Dad", color: "#2D5A3D" },
  { name: "Mom", color: "#C4704B" },
  { name: "Grampa", color: "#D4A84B" },
  { name: "Auntie", color: "#7B5EA7" },
  { name: "Nana", color: "#D4704B" },
  { name: "Grammy", color: "#3A7A52" },
  { name: "Papa", color: "#6B5744" },
  { name: "Uncle", color: "#2D7A8A" },
];

const DISPLAY_MS = 2500;
const TRANSITION_MS = 200;

interface RotatingNameProps {
  className?: string;
}

export default function RotatingName({ className = "" }: RotatingNameProps) {
  const [index, setIndex] = useState(0);
  const [isTransitioning, setIsTransitioning] = useState(false);
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);
  const nextIndexRef = useRef(1);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    setPrefersReducedMotion(mq.matches);
    const handler = (e: MediaQueryListEvent) => setPrefersReducedMotion(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  useEffect(() => {
    if (prefersReducedMotion) return;

    const interval = setInterval(() => {
      setIsTransitioning(true);
      setTimeout(() => {
        setIndex(nextIndexRef.current);
        nextIndexRef.current = (nextIndexRef.current + 1) % NAMES.length;
        setIsTransitioning(false);
      }, TRANSITION_MS);
    }, DISPLAY_MS + TRANSITION_MS);

    return () => clearInterval(interval);
  }, [prefersReducedMotion]);

  if (prefersReducedMotion) {
    return (
      <span className={className} style={{ color: NAMES[0].color }}>
        Dad
      </span>
    );
  }

  const current = NAMES[index];

  return (
    <span
      className={`inline-flex items-baseline overflow-hidden relative align-baseline ${className}`}
      style={{ height: "1.2em" }}
      aria-live="polite"
    >
      <span
        className={isTransitioning ? "animate-name-slide-out" : "animate-name-slide-in"}
        style={{
          color: current.color,
          display: "inline-block",
          position: "relative",
        }}
      >
        {current.name}
      </span>
    </span>
  );
}
