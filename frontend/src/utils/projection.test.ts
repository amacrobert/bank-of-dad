import { describe, it, expect } from "vitest";
import { calculateProjection } from "./projection";
import { ProjectionConfig } from "../types";

function makeConfig(overrides: Partial<ProjectionConfig> = {}): ProjectionConfig {
  return {
    currentBalanceCents: 10000, // $100
    interestRateBps: 0,
    interestFrequency: null,
    allowanceAmountCents: 0,
    allowanceFrequency: null,
    scenario: {
      weeklySpendingCents: 0,
      weeklySavingsCents: 0,
      oneTimeDepositCents: 0,
      oneTimeWithdrawalCents: 0,
      horizonMonths: 12,
    },
    ...overrides,
  };
}

describe("calculateProjection", () => {
  // (a) Allowance-only linear growth
  describe("allowance-only linear growth", () => {
    it("adds weekly allowance each week for 1 year", () => {
      const config = makeConfig({
        currentBalanceCents: 10000,
        allowanceAmountCents: 1000, // $10/week
        allowanceFrequency: "weekly",
      });
      const result = calculateProjection(config);

      // 52 weeks * $10 = $520 in allowance
      expect(result.totalAllowanceCents).toBe(52 * 1000);
      // Final = $100 + $520 = $620
      expect(result.finalBalanceCents).toBe(10000 + 52 * 1000);
      expect(result.totalInterestCents).toBe(0);
      expect(result.startingBalanceCents).toBe(10000);
    });

    it("adds biweekly allowance every 2 weeks", () => {
      const config = makeConfig({
        currentBalanceCents: 10000,
        allowanceAmountCents: 2000, // $20 biweekly
        allowanceFrequency: "biweekly",
      });
      const result = calculateProjection(config);

      // 52 weeks / 2 = 26 payments * $20 = $520
      expect(result.totalAllowanceCents).toBe(26 * 2000);
      expect(result.finalBalanceCents).toBe(10000 + 26 * 2000);
    });

    it("adds monthly allowance approximately every 4.33 weeks", () => {
      const config = makeConfig({
        currentBalanceCents: 0,
        allowanceAmountCents: 5000, // $50/month
        allowanceFrequency: "monthly",
      });
      const result = calculateProjection(config);

      // 12 monthly payments * $50 = $600
      expect(result.totalAllowanceCents).toBe(12 * 5000);
      expect(result.finalBalanceCents).toBe(12 * 5000);
    });
  });

  // (b) Interest-only compound growth
  describe("interest-only compound growth", () => {
    it("compounds weekly interest over 1 year", () => {
      const config = makeConfig({
        currentBalanceCents: 100000, // $1000
        interestRateBps: 500, // 5%
        interestFrequency: "weekly",
      });
      const result = calculateProjection(config);

      // Weekly rate = 5% / 52 = ~0.0961538%
      // After 52 weeks of compounding: $1000 * (1 + 0.05/52)^52 ≈ $1051.25
      expect(result.totalInterestCents).toBeGreaterThan(0);
      expect(result.finalBalanceCents).toBeGreaterThan(100000);
      // Verify approximately correct (within $1)
      expect(result.finalBalanceCents).toBeCloseTo(105125, -2);
      expect(result.totalAllowanceCents).toBe(0);
    });

    it("compounds biweekly interest", () => {
      const config = makeConfig({
        currentBalanceCents: 100000, // $1000
        interestRateBps: 500, // 5%
        interestFrequency: "biweekly",
      });
      const result = calculateProjection(config);

      // Biweekly rate = 5% / 26
      // After 26 compounding periods: $1000 * (1 + 0.05/26)^26 ≈ $1051.19
      expect(result.totalInterestCents).toBeGreaterThan(0);
      expect(result.finalBalanceCents).toBeGreaterThan(100000);
    });

    it("compounds monthly interest", () => {
      const config = makeConfig({
        currentBalanceCents: 100000, // $1000
        interestRateBps: 500, // 5%
        interestFrequency: "monthly",
      });
      const result = calculateProjection(config);

      // Monthly rate = 5% / 12
      // After 12 months: $1000 * (1 + 0.05/12)^12 ≈ $1051.16
      expect(result.totalInterestCents).toBeGreaterThan(0);
      expect(result.finalBalanceCents).toBeGreaterThan(100000);
    });
  });

  // (c) Combined allowance + interest
  describe("combined allowance and interest", () => {
    it("applies both allowance deposits and interest compounding", () => {
      const config = makeConfig({
        currentBalanceCents: 10000, // $100
        interestRateBps: 500, // 5%
        interestFrequency: "weekly",
        allowanceAmountCents: 1000, // $10/week
        allowanceFrequency: "weekly",
      });
      const result = calculateProjection(config);

      // Should have both interest and allowance
      expect(result.totalInterestCents).toBeGreaterThan(0);
      expect(result.totalAllowanceCents).toBe(52 * 1000);
      // Final should be more than allowance-only ($100 + $520 = $620)
      expect(result.finalBalanceCents).toBeGreaterThan(62000);
    });
  });

  // (d) Different frequencies
  describe("frequency combinations", () => {
    it("handles weekly allowance with monthly interest", () => {
      const config = makeConfig({
        currentBalanceCents: 10000,
        interestRateBps: 1000, // 10%
        interestFrequency: "monthly",
        allowanceAmountCents: 500, // $5/week
        allowanceFrequency: "weekly",
      });
      const result = calculateProjection(config);

      expect(result.totalAllowanceCents).toBe(52 * 500);
      expect(result.totalInterestCents).toBeGreaterThan(0);
    });
  });

  // (e) Zero balance with no schedules (flat line)
  describe("zero balance with no schedules", () => {
    it("returns flat line at $0", () => {
      const config = makeConfig({
        currentBalanceCents: 0,
      });
      const result = calculateProjection(config);

      expect(result.finalBalanceCents).toBe(0);
      expect(result.totalInterestCents).toBe(0);
      expect(result.totalAllowanceCents).toBe(0);
      expect(result.startingBalanceCents).toBe(0);
      expect(result.depletionWeek).toBeNull();
      // All data points should be 0
      result.dataPoints.forEach((dp) => {
        expect(dp.balanceCents).toBe(0);
      });
    });

    it("returns flat line at current balance when no growth sources", () => {
      const config = makeConfig({
        currentBalanceCents: 5000, // $50
      });
      const result = calculateProjection(config);

      expect(result.finalBalanceCents).toBe(5000);
      result.dataPoints.forEach((dp) => {
        expect(dp.balanceCents).toBe(5000);
      });
    });
  });

  // (f) Balance floor at $0
  describe("balance floor at $0", () => {
    it("floors balance at 0 when spending exceeds income", () => {
      const config = makeConfig({
        currentBalanceCents: 5000, // $50
        scenario: {
          weeklySpendingCents: 2000, // $20/week
          weeklySavingsCents: 0,
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      expect(result.finalBalanceCents).toBe(0);
      // No data point should be negative
      result.dataPoints.forEach((dp) => {
        expect(dp.balanceCents).toBeGreaterThanOrEqual(0);
      });
    });
  });

  // (g) Depletion week detection
  describe("depletion week detection", () => {
    it("detects when balance reaches 0", () => {
      const config = makeConfig({
        currentBalanceCents: 5000, // $50
        scenario: {
          weeklySpendingCents: 1000, // $10/week
          weeklySavingsCents: 0,
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      // $50 / $10/week = 5 weeks
      expect(result.depletionWeek).toBe(5);
    });

    it("returns null when balance never depletes", () => {
      const config = makeConfig({
        currentBalanceCents: 10000,
        allowanceAmountCents: 1000,
        allowanceFrequency: "weekly",
      });
      const result = calculateProjection(config);

      expect(result.depletionWeek).toBeNull();
    });

    it("detects depletion with allowance partially offsetting spending", () => {
      const config = makeConfig({
        currentBalanceCents: 10000, // $100
        allowanceAmountCents: 500, // $5/week
        allowanceFrequency: "weekly",
        scenario: {
          weeklySpendingCents: 1000, // $10/week (net -$5/week)
          weeklySavingsCents: 0,
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      // Net drain is $5/week, $100 / $5 = 20 weeks
      expect(result.depletionWeek).toBe(20);
    });
  });

  // (h) One-time deposit/withdrawal adjustments
  describe("one-time adjustments", () => {
    it("adds one-time deposit to starting balance", () => {
      const config = makeConfig({
        currentBalanceCents: 10000, // $100
        scenario: {
          weeklySpendingCents: 0,
          weeklySavingsCents: 0,
          oneTimeDepositCents: 5000, // $50
          oneTimeWithdrawalCents: 0,
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      expect(result.startingBalanceCents).toBe(15000);
      expect(result.finalBalanceCents).toBe(15000);
    });

    it("subtracts one-time withdrawal from starting balance", () => {
      const config = makeConfig({
        currentBalanceCents: 10000, // $100
        scenario: {
          weeklySpendingCents: 0,
          weeklySavingsCents: 0,
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 3000, // $30
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      expect(result.startingBalanceCents).toBe(7000);
      expect(result.finalBalanceCents).toBe(7000);
    });

    it("clamps withdrawal to current balance (no negative start)", () => {
      const config = makeConfig({
        currentBalanceCents: 5000, // $50
        scenario: {
          weeklySpendingCents: 0,
          weeklySavingsCents: 0,
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 10000, // $100 > $50 balance
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      expect(result.startingBalanceCents).toBe(0);
      expect(result.finalBalanceCents).toBe(0);
    });
  });

  // (i) Paused schedules excluded
  describe("paused schedules excluded", () => {
    it("excludes allowance when frequency is null (paused/none)", () => {
      const config = makeConfig({
        currentBalanceCents: 10000,
        allowanceAmountCents: 1000,
        allowanceFrequency: null, // paused
      });
      const result = calculateProjection(config);

      expect(result.totalAllowanceCents).toBe(0);
      expect(result.finalBalanceCents).toBe(10000);
    });

    it("excludes interest when frequency is null (paused/none)", () => {
      const config = makeConfig({
        currentBalanceCents: 10000,
        interestRateBps: 500,
        interestFrequency: null, // paused
      });
      const result = calculateProjection(config);

      expect(result.totalInterestCents).toBe(0);
      expect(result.finalBalanceCents).toBe(10000);
    });
  });

  // Time horizon tests
  describe("time horizons", () => {
    it("generates correct number of data points for 3-month horizon", () => {
      const config = makeConfig({
        scenario: {
          weeklySpendingCents: 0,
          weeklySavingsCents: 0,
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 3,
        },
      });
      const result = calculateProjection(config);

      // 3 months ≈ 13 weeks + 1 for week 0
      expect(result.dataPoints.length).toBe(14);
    });

    it("generates correct number of data points for 5-year horizon", () => {
      const config = makeConfig({
        scenario: {
          weeklySpendingCents: 0,
          weeklySavingsCents: 0,
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 60,
        },
      });
      const result = calculateProjection(config);

      // 60 months ≈ 260 weeks + 1 for week 0
      expect(result.dataPoints.length).toBe(261);
    });

    it("starts data points at week 0 with current balance", () => {
      const config = makeConfig({
        currentBalanceCents: 10000,
      });
      const result = calculateProjection(config);

      expect(result.dataPoints[0].weekIndex).toBe(0);
      expect(result.dataPoints[0].balanceCents).toBe(10000);
    });
  });

  // Spending tracking
  describe("spending tracking", () => {
    it("tracks total spending correctly", () => {
      const config = makeConfig({
        currentBalanceCents: 100000, // $1000 so we don't deplete
        scenario: {
          weeklySpendingCents: 500, // $5/week
          weeklySavingsCents: 0,
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      expect(result.totalSpendingCents).toBe(52 * 500);
    });
  });

  // Component breakdown sums to total
  describe("component breakdown integrity", () => {
    it("components sum to final balance", () => {
      const config = makeConfig({
        currentBalanceCents: 10000, // $100
        interestRateBps: 500, // 5%
        interestFrequency: "weekly",
        allowanceAmountCents: 1000, // $10/week
        allowanceFrequency: "weekly",
        scenario: {
          weeklySpendingCents: 300, // $3/week
          weeklySavingsCents: 0,
          oneTimeDepositCents: 2000, // $20
          oneTimeWithdrawalCents: 1000, // $10
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      // Final = starting + interest + allowance + savings - spending
      const expectedFinal =
        result.startingBalanceCents +
        result.totalInterestCents +
        result.totalAllowanceCents +
        result.totalSavingsCents -
        result.totalSpendingCents;
      expect(result.finalBalanceCents).toBe(expectedFinal);
    });
  });

  // (k) Weekly savings support
  describe("weekly savings", () => {
    it("adds weekly savings to balance each week", () => {
      const config = makeConfig({
        currentBalanceCents: 10000, // $100
        scenario: {
          weeklySpendingCents: 0,
          weeklySavingsCents: 500, // $5/week
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      // 52 weeks * $5 = $260 in savings
      expect(result.totalSavingsCents).toBe(52 * 500);
      // Final = $100 + $260 = $360
      expect(result.finalBalanceCents).toBe(10000 + 52 * 500);
    });

    it("stacks savings with allowance deposits", () => {
      const config = makeConfig({
        currentBalanceCents: 10000, // $100
        allowanceAmountCents: 1000, // $10/week
        allowanceFrequency: "weekly",
        scenario: {
          weeklySpendingCents: 0,
          weeklySavingsCents: 500, // $5/week extra
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      expect(result.totalAllowanceCents).toBe(52 * 1000);
      expect(result.totalSavingsCents).toBe(52 * 500);
      // Final = $100 + $520 (allowance) + $260 (savings)
      expect(result.finalBalanceCents).toBe(10000 + 52 * 1000 + 52 * 500);
    });

    it("compounds correctly with interest and savings", () => {
      const config = makeConfig({
        currentBalanceCents: 100000, // $1000
        interestRateBps: 500, // 5%
        interestFrequency: "weekly",
        scenario: {
          weeklySpendingCents: 0,
          weeklySavingsCents: 1000, // $10/week
          oneTimeDepositCents: 0,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      // Should have both interest and savings
      expect(result.totalInterestCents).toBeGreaterThan(0);
      expect(result.totalSavingsCents).toBe(52 * 1000);
      // Final should be more than just savings ($1000 + $520 = $1520)
      expect(result.finalBalanceCents).toBeGreaterThan(152000);
    });

    it("includes savings in component breakdown integrity", () => {
      const config = makeConfig({
        currentBalanceCents: 10000,
        interestRateBps: 500,
        interestFrequency: "weekly",
        allowanceAmountCents: 1000,
        allowanceFrequency: "weekly",
        scenario: {
          weeklySpendingCents: 0,
          weeklySavingsCents: 200, // $2/week
          oneTimeDepositCents: 2000,
          oneTimeWithdrawalCents: 0,
          horizonMonths: 12,
        },
      });
      const result = calculateProjection(config);

      // Final = starting + interest + allowance + savings - spending
      const expectedFinal =
        result.startingBalanceCents +
        result.totalInterestCents +
        result.totalAllowanceCents +
        result.totalSavingsCents -
        result.totalSpendingCents;
      expect(result.finalBalanceCents).toBe(expectedFinal);
    });
  });
});
