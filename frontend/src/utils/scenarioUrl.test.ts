import { describe, it, expect } from "vitest";
import { serializeScenarios, deserializeScenarios } from "./scenarioUrl";
import { ScenarioConfig, SCENARIO_COLORS } from "../types";

function makeScenario(overrides: Partial<ScenarioConfig> = {}): ScenarioConfig {
  return {
    id: "s0",
    weeklyAmountCents: 0,
    weeklyDirection: "spending",
    oneTimeAmountCents: 0,
    oneTimeDirection: "deposit",
    color: "#2563eb",
    ...overrides,
  };
}

describe("scenarioUrl", () => {
  it("round-trip: serialize then deserialize returns equivalent scenarios", () => {
    const scenarios = [
      makeScenario({ weeklyAmountCents: 500, weeklyDirection: "saving" }),
      makeScenario({
        id: "s1",
        oneTimeAmountCents: 10000,
        oneTimeDirection: "withdrawal",
        color: "#dc2626",
      }),
    ];
    const qs = serializeScenarios(scenarios, 24);
    const params = new URLSearchParams(qs);
    const result = deserializeScenarios(params);

    expect(result).not.toBeNull();
    expect(result!.scenarios).toHaveLength(2);

    // Compare value fields, not id/color (those are reassigned on deserialize)
    expect(result!.scenarios[0].weeklyAmountCents).toBe(500);
    expect(result!.scenarios[0].weeklyDirection).toBe("saving");
    expect(result!.scenarios[0].oneTimeAmountCents).toBe(0);
    expect(result!.scenarios[0].oneTimeDirection).toBe("deposit");

    expect(result!.scenarios[1].weeklyAmountCents).toBe(0);
    expect(result!.scenarios[1].weeklyDirection).toBe("spending");
    expect(result!.scenarios[1].oneTimeAmountCents).toBe(10000);
    expect(result!.scenarios[1].oneTimeDirection).toBe("withdrawal");
  });

  it("compact output: serialized string contains scenarios= param with base64 value", () => {
    const scenarios = [makeScenario({ weeklyAmountCents: 100 })];
    const qs = serializeScenarios(scenarios, 12);

    expect(qs).toContain("scenarios=");
    const params = new URLSearchParams(qs);
    const encoded = params.get("scenarios");
    expect(encoded).toBeTruthy();
    // URL-safe base64 should not contain +, /, or = padding (uses - _ instead)
    expect(encoded).toMatch(/^[A-Za-z0-9_-]+$/);
  });

  it("missing param: deserialize with no scenarios param returns null", () => {
    const params = new URLSearchParams("h=12");
    expect(deserializeScenarios(params)).toBeNull();
  });

  it("malformed base64: deserialize with garbage scenarios value returns null", () => {
    const params = new URLSearchParams("scenarios=!!!not-valid-base64&h=12");
    expect(deserializeScenarios(params)).toBeNull();
  });

  it("invalid JSON: deserialize with valid base64 but invalid JSON structure returns null", () => {
    // Encode a string that is valid base64 but not valid scenario JSON
    const raw = btoa("this is not json");
    // Convert standard base64 to URL-safe
    const urlSafe = raw
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=+$/, "");
    const params = new URLSearchParams(`scenarios=${urlSafe}&h=12`);
    expect(deserializeScenarios(params)).toBeNull();
  });

  it("horizon round-trip: horizonMonths value survives serialize/deserialize", () => {
    const scenarios = [makeScenario()];
    const qs = serializeScenarios(scenarios, 36);
    const params = new URLSearchParams(qs);
    const result = deserializeScenarios(params);

    expect(result).not.toBeNull();
    expect(result!.horizonMonths).toBe(36);
  });

  it("handles 1 scenario", () => {
    const scenarios = [
      makeScenario({ weeklyAmountCents: 250, weeklyDirection: "saving" }),
    ];
    const qs = serializeScenarios(scenarios, 12);
    const params = new URLSearchParams(qs);
    const result = deserializeScenarios(params);

    expect(result).not.toBeNull();
    expect(result!.scenarios).toHaveLength(1);
    expect(result!.scenarios[0].weeklyAmountCents).toBe(250);
    expect(result!.scenarios[0].weeklyDirection).toBe("saving");
  });

  it("handles 5 scenarios", () => {
    const scenarios = SCENARIO_COLORS.map((color, i) =>
      makeScenario({
        id: `s${i}`,
        weeklyAmountCents: (i + 1) * 100,
        color,
      })
    );
    const qs = serializeScenarios(scenarios, 12);
    const params = new URLSearchParams(qs);
    const result = deserializeScenarios(params);

    expect(result).not.toBeNull();
    expect(result!.scenarios).toHaveLength(5);
    result!.scenarios.forEach((s: ScenarioConfig, i: number) => {
      expect(s.weeklyAmountCents).toBe((i + 1) * 100);
    });
  });

  it("default horizon: if h param is missing, uses default of 12", () => {
    const scenarios = [makeScenario()];
    const qs = serializeScenarios(scenarios, 12);
    const params = new URLSearchParams(qs);
    // Remove the h param to simulate it being missing
    params.delete("h");
    const result = deserializeScenarios(params);

    expect(result).not.toBeNull();
    expect(result!.horizonMonths).toBe(12);
  });
});
