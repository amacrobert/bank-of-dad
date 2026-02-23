import { ScenarioConfig, SCENARIO_COLORS } from "../types";
import type { WeeklyDirection, OneTimeDirection } from "../types";

interface CompactScenario {
  w: number;
  wd: "s" | "v"; // spending | saving
  o: number;
  od: "d" | "w"; // deposit | withdrawal
}

function toCompact(sc: ScenarioConfig): CompactScenario {
  return {
    w: sc.weeklyAmountCents,
    wd: sc.weeklyDirection === "saving" ? "v" : "s",
    o: sc.oneTimeAmountCents,
    od: sc.oneTimeDirection === "withdrawal" ? "w" : "d",
  };
}

function fromCompact(c: CompactScenario, index: number): ScenarioConfig {
  return {
    id: `s${index}`,
    weeklyAmountCents: c.w,
    weeklyDirection: (c.wd === "v" ? "saving" : "spending") as WeeklyDirection,
    oneTimeAmountCents: c.o,
    oneTimeDirection: (c.od === "w" ? "withdrawal" : "deposit") as OneTimeDirection,
    color: SCENARIO_COLORS[index % SCENARIO_COLORS.length],
  };
}

function toUrlSafeBase64(str: string): string {
  return btoa(str)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

function fromUrlSafeBase64(encoded: string): string {
  // Restore standard base64
  let base64 = encoded.replace(/-/g, "+").replace(/_/g, "/");
  // Add padding
  while (base64.length % 4 !== 0) {
    base64 += "=";
  }
  return atob(base64);
}

export function serializeScenarios(scenarios: ScenarioConfig[], horizonMonths: number): string {
  const compact = scenarios.map(toCompact);
  const json = JSON.stringify(compact);
  const encoded = toUrlSafeBase64(json);
  const params = new URLSearchParams();
  params.set("scenarios", encoded);
  params.set("h", String(horizonMonths));
  return params.toString();
}

export function deserializeScenarios(
  params: URLSearchParams
): { scenarios: ScenarioConfig[]; horizonMonths: number } | null {
  const encoded = params.get("scenarios");
  if (!encoded) return null;

  try {
    const json = fromUrlSafeBase64(encoded);
    const compact = JSON.parse(json);
    if (!Array.isArray(compact)) return null;

    const scenarios = compact.map((c: CompactScenario, i: number) => fromCompact(c, i));
    const h = params.get("h");
    const horizonMonths = h ? parseInt(h, 10) : 12;

    return {
      scenarios,
      horizonMonths: isNaN(horizonMonths) ? 12 : horizonMonths,
    };
  } catch {
    return null;
  }
}
