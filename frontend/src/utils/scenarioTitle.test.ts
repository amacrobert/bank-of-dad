import { describe, it, expect } from "vitest";
import { generateScenarioTitle } from "./scenarioTitle";
import { ScenarioTitleContext } from "../types";

function ctx(
  overrides: Partial<ScenarioTitleContext> = {}
): ScenarioTitleContext {
  return {
    hasAllowance: true,
    weeklyAllowanceCents: 2000,
    weeklyAmountCents: 0,
    weeklyDirection: "spending",
    oneTimeAmountCents: 0,
    oneTimeDirection: "deposit",
    ...overrides,
  };
}

describe("generateScenarioTitle", () => {
  describe("has allowance, no one-time", () => {
    it("spending=0 → save all of allowance", () => {
      expect(generateScenarioTitle(ctx())).toBe(
        "If I save **all** of my $20 allowance"
      );
    });

    it("spending == allowance → save none of allowance", () => {
      expect(
        generateScenarioTitle(
          ctx({ weeklyAmountCents: 2000, weeklyDirection: "spending" })
        )
      ).toBe("If I save **none** of my $20 allowance");
    });

    it("0 < spending < allowance → save partial amount", () => {
      expect(
        generateScenarioTitle(
          ctx({ weeklyAmountCents: 1000, weeklyDirection: "spending" })
        )
      ).toBe("If I save **$10** per week from my $20 allowance");
    });

    it("saving > 0 → save all plus additional", () => {
      expect(
        generateScenarioTitle(
          ctx({ weeklyAmountCents: 500, weeklyDirection: "saving" })
        )
      ).toBe(
        "If I save **all** of my $20 allowance plus an additional **$5** per week"
      );
    });
  });

  describe("has allowance + deposit", () => {
    it("spending=0 → save all, and deposit", () => {
      expect(
        generateScenarioTitle(
          ctx({ oneTimeAmountCents: 5000, oneTimeDirection: "deposit" })
        )
      ).toBe(
        "If I save **all** of my $20 allowance, and deposit **$50** now"
      );
    });

    it("spending == allowance → save none, and deposit", () => {
      expect(
        generateScenarioTitle(
          ctx({
            weeklyAmountCents: 2000,
            weeklyDirection: "spending",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "deposit",
          })
        )
      ).toBe(
        "If I save **none** of my $20 allowance, and deposit **$50** now"
      );
    });

    it("0 < spending < allowance → save partial, and deposit", () => {
      expect(
        generateScenarioTitle(
          ctx({
            weeklyAmountCents: 1000,
            weeklyDirection: "spending",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "deposit",
          })
        )
      ).toBe(
        "If I save **$10** per week from my $20 allowance, and deposit **$50** now"
      );
    });

    it("saving > 0 → save all plus additional, and deposit", () => {
      expect(
        generateScenarioTitle(
          ctx({
            weeklyAmountCents: 500,
            weeklyDirection: "saving",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "deposit",
          })
        )
      ).toBe(
        "If I save **all** of my $20 allowance plus an additional **$5** per week, and deposit **$50** now"
      );
    });
  });

  describe("has allowance + withdrawal", () => {
    it("spending=0 → save all, but withdraw", () => {
      expect(
        generateScenarioTitle(
          ctx({ oneTimeAmountCents: 5000, oneTimeDirection: "withdrawal" })
        )
      ).toBe(
        "If I save **all** of my $20 allowance, but withdraw **$50** now"
      );
    });

    it("spending == allowance → save none, but withdraw", () => {
      expect(
        generateScenarioTitle(
          ctx({
            weeklyAmountCents: 2000,
            weeklyDirection: "spending",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "withdrawal",
          })
        )
      ).toBe(
        "If I save **none** of my $20 allowance, but withdraw **$50** now"
      );
    });

    it("0 < spending < allowance → save partial, but withdraw", () => {
      expect(
        generateScenarioTitle(
          ctx({
            weeklyAmountCents: 1000,
            weeklyDirection: "spending",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "withdrawal",
          })
        )
      ).toBe(
        "If I save **$10** per week from my $20 allowance, but withdraw **$50** now"
      );
    });

    it("saving > 0 → save all plus additional, but withdraw", () => {
      expect(
        generateScenarioTitle(
          ctx({
            weeklyAmountCents: 500,
            weeklyDirection: "saving",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "withdrawal",
          })
        )
      ).toBe(
        "If I save **all** of my $20 allowance plus an additional **$5** per week, but withdraw **$50** now"
      );
    });
  });

  describe("no allowance, no one-time", () => {
    it("spending=0 → don't do anything", () => {
      expect(generateScenarioTitle(ctx({ hasAllowance: false }))).toBe(
        "If I don't do anything"
      );
    });

    it("spending > 0 → spend per week", () => {
      expect(
        generateScenarioTitle(
          ctx({
            hasAllowance: false,
            weeklyAmountCents: 500,
            weeklyDirection: "spending",
          })
        )
      ).toBe("If I spend **$5** per week");
    });

    it("saving > 0 → save per week", () => {
      expect(
        generateScenarioTitle(
          ctx({
            hasAllowance: false,
            weeklyAmountCents: 1000,
            weeklyDirection: "saving",
          })
        )
      ).toBe("If I save **$10** per week");
    });
  });

  describe("no allowance + deposit", () => {
    it("spending=0 → deposit only", () => {
      expect(
        generateScenarioTitle(
          ctx({
            hasAllowance: false,
            oneTimeAmountCents: 5000,
            oneTimeDirection: "deposit",
          })
        )
      ).toBe("If I deposit **$50** now");
    });

    it("spending > 0 → deposit and spend", () => {
      expect(
        generateScenarioTitle(
          ctx({
            hasAllowance: false,
            weeklyAmountCents: 500,
            weeklyDirection: "spending",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "deposit",
          })
        )
      ).toBe("If I deposit **$50** now and spend **$5** per week");
    });

    it("saving > 0 → save and deposit", () => {
      expect(
        generateScenarioTitle(
          ctx({
            hasAllowance: false,
            weeklyAmountCents: 1000,
            weeklyDirection: "saving",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "deposit",
          })
        )
      ).toBe("If I save **$10** per week, and deposit **$50** now");
    });
  });

  describe("no allowance + withdrawal", () => {
    it("spending=0 → withdraw only", () => {
      expect(
        generateScenarioTitle(
          ctx({
            hasAllowance: false,
            oneTimeAmountCents: 5000,
            oneTimeDirection: "withdrawal",
          })
        )
      ).toBe("If I withdraw **$50** now");
    });

    it("spending > 0 → spend and withdraw", () => {
      expect(
        generateScenarioTitle(
          ctx({
            hasAllowance: false,
            weeklyAmountCents: 500,
            weeklyDirection: "spending",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "withdrawal",
          })
        )
      ).toBe("If I spend **$5** per week, and withdraw **$50** now");
    });

    it("saving > 0 → save but withdraw", () => {
      expect(
        generateScenarioTitle(
          ctx({
            hasAllowance: false,
            weeklyAmountCents: 1000,
            weeklyDirection: "saving",
            oneTimeAmountCents: 5000,
            oneTimeDirection: "withdrawal",
          })
        )
      ).toBe("If I save **$10** per week, but withdraw **$50** now");
    });
  });
});
