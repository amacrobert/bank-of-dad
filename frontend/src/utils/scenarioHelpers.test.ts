import { describe, it, expect } from "vitest";
import { mapScenarioConfigToInputs, buildDefaultScenarios } from "./scenarioHelpers";
import { ScenarioConfig } from "../types";

describe("mapScenarioConfigToInputs", () => {
  it("maps spending direction to weeklySpendingCents", () => {
    const config: ScenarioConfig = {
      id: "s0",
      weeklyAmountCents: 500,
      weeklyDirection: "spending",
      oneTimeAmountCents: 0,
      oneTimeDirection: "deposit",
      color: "#2563eb",
    };
    const result = mapScenarioConfigToInputs(config, 12);

    expect(result.weeklySpendingCents).toBe(500);
    expect(result.weeklySavingsCents).toBe(0);
  });

  it("maps saving direction to weeklySavingsCents", () => {
    const config: ScenarioConfig = {
      id: "s0",
      weeklyAmountCents: 300,
      weeklyDirection: "saving",
      oneTimeAmountCents: 0,
      oneTimeDirection: "deposit",
      color: "#2563eb",
    };
    const result = mapScenarioConfigToInputs(config, 12);

    expect(result.weeklySpendingCents).toBe(0);
    expect(result.weeklySavingsCents).toBe(300);
  });

  it("maps deposit direction to oneTimeDepositCents", () => {
    const config: ScenarioConfig = {
      id: "s0",
      weeklyAmountCents: 0,
      weeklyDirection: "spending",
      oneTimeAmountCents: 5000,
      oneTimeDirection: "deposit",
      color: "#2563eb",
    };
    const result = mapScenarioConfigToInputs(config, 12);

    expect(result.oneTimeDepositCents).toBe(5000);
    expect(result.oneTimeWithdrawalCents).toBe(0);
  });

  it("maps withdrawal direction to oneTimeWithdrawalCents", () => {
    const config: ScenarioConfig = {
      id: "s0",
      weeklyAmountCents: 0,
      weeklyDirection: "spending",
      oneTimeAmountCents: 3000,
      oneTimeDirection: "withdrawal",
      color: "#2563eb",
    };
    const result = mapScenarioConfigToInputs(config, 12);

    expect(result.oneTimeDepositCents).toBe(0);
    expect(result.oneTimeWithdrawalCents).toBe(3000);
  });

  it("passes horizonMonths through", () => {
    const config: ScenarioConfig = {
      id: "s0",
      weeklyAmountCents: 0,
      weeklyDirection: "spending",
      oneTimeAmountCents: 0,
      oneTimeDirection: "deposit",
      color: "#2563eb",
    };
    const result = mapScenarioConfigToInputs(config, 24);

    expect(result.horizonMonths).toBe(24);
  });
});

describe("buildDefaultScenarios", () => {
  it("returns two scenarios when child has allowance", () => {
    const result = buildDefaultScenarios(2000, "weekly");

    expect(result).toHaveLength(2);
    // First: save all (spending = 0)
    expect(result[0].weeklyAmountCents).toBe(0);
    expect(result[0].weeklyDirection).toBe("spending");
    // Second: spend all (spending = 100% weekly allowance)
    expect(result[1].weeklyAmountCents).toBe(2000);
    expect(result[1].weeklyDirection).toBe("spending");
  });

  it("converts biweekly allowance to weekly for defaults", () => {
    const result = buildDefaultScenarios(2000, "biweekly");

    expect(result).toHaveLength(2);
    // $20 biweekly = $10/week
    expect(result[1].weeklyAmountCents).toBe(1000);
    expect(result[1].weeklyDirection).toBe("spending");
  });

  it("converts monthly allowance to weekly for defaults", () => {
    const result = buildDefaultScenarios(5000, "monthly");

    expect(result).toHaveLength(2);
    // $50 monthly ÷ 4.333 weeks ≈ $11.54/week = 1154 cents
    expect(result[1].weeklyAmountCents).toBe(Math.round(5000 / (52 / 12)));
    expect(result[1].weeklyDirection).toBe("spending");
  });

  it("returns two scenarios when child has no allowance", () => {
    const result = buildDefaultScenarios(0, null);

    expect(result).toHaveLength(2);
    // First: no spending
    expect(result[0].weeklyAmountCents).toBe(0);
    expect(result[0].weeklyDirection).toBe("spending");
    // Second: $5/week spending
    expect(result[1].weeklyAmountCents).toBe(500);
    expect(result[1].weeklyDirection).toBe("spending");
  });

  it("assigns distinct colors from palette", () => {
    const result = buildDefaultScenarios(1000, "weekly");

    expect(result[0].color).not.toBe(result[1].color);
  });

  it("assigns unique ids", () => {
    const result = buildDefaultScenarios(1000, "weekly");

    expect(result[0].id).not.toBe(result[1].id);
  });

  it("defaults one-time amounts to 0 deposit", () => {
    const result = buildDefaultScenarios(1000, "weekly");

    for (const s of result) {
      expect(s.oneTimeAmountCents).toBe(0);
      expect(s.oneTimeDirection).toBe("deposit");
    }
  });
});
