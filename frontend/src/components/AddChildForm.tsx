import { useState } from "react";
import { post, ApiRequestError, deposit, setChildAllowance, setInterest } from "../api";
import { ChildCreateResponse } from "../types";
import Input from "./ui/Input";
import Button from "./ui/Button";
import AvatarPicker from "./AvatarPicker";
import { CheckCircle } from "lucide-react";

interface AddChildFormProps {
  onChildAdded: () => void;
  onCancel?: () => void;
}

export default function AddChildForm({ onChildAdded, onCancel }: AddChildFormProps) {
  const [firstName, setFirstName] = useState("");
  const [password, setPassword] = useState("");
  const [avatar, setAvatar] = useState<string | null>(null);
  const [initialDeposit, setInitialDeposit] = useState("");
  const [weeklyAllowance, setWeeklyAllowance] = useState("");
  const [annualInterest, setAnnualInterest] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [setupWarning, setSetupWarning] = useState<string | null>(null);
  const [created, setCreated] = useState<ChildCreateResponse | null>(null);
  const [setupSummary, setSetupSummary] = useState<{
    depositAmount?: number;
    allowanceSet?: boolean;
    interestSet?: boolean;
  } | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setCreated(null);
    setSetupWarning(null);
    setSetupSummary(null);

    try {
      const result = await post<ChildCreateResponse>("/children", {
        first_name: firstName,
        password,
        avatar: avatar || undefined,
      });

      // Post-creation setup
      const warnings: string[] = [];
      const summary: typeof setupSummary = {};

      // Initial deposit
      const depositAmount = parseFloat(initialDeposit);
      if (!isNaN(depositAmount) && depositAmount > 0) {
        try {
          const depositCents = Math.round(depositAmount * 100);
          await deposit(result.id, { amount_cents: depositCents, note: "Initial deposit" });
          summary.depositAmount = depositAmount;
        } catch {
          warnings.push("initial deposit");
        }
      }

      // Weekly allowance
      const allowanceAmount = parseFloat(weeklyAllowance);
      if (!isNaN(allowanceAmount) && allowanceAmount > 0) {
        try {
          const allowanceCents = Math.round(allowanceAmount * 100);
          await setChildAllowance(result.id, {
            amount_cents: allowanceCents,
            frequency: "weekly",
            day_of_week: new Date().getDay(),
            note: "Weekly allowance",
          });
          summary.allowanceSet = true;
        } catch {
          warnings.push("weekly allowance");
        }
      }

      // Annual interest
      const interestRate = parseFloat(annualInterest);
      if (!isNaN(interestRate) && interestRate > 0) {
        try {
          const bps = Math.round(interestRate * 100);
          await setInterest(result.id, {
            interest_rate_bps: bps,
            frequency: "monthly",
            day_of_month: 1,
          });
          summary.interestSet = true;
        } catch {
          warnings.push("annual interest");
        }
      }

      if (warnings.length > 0) {
        setSetupWarning(`Child created, but failed to set up: ${warnings.join(", ")}. You can configure these from the child's settings.`);
      }

      setCreated(result);
      setSetupSummary(summary);
      setFirstName("");
      setPassword("");
      setAvatar(null);
      setInitialDeposit("");
      setWeeklyAllowance("");
      setAnnualInterest("");
      onChildAdded();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <h3 className="text-base font-bold text-bark mb-3">Add a new child</h3>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="First Name"
          id="child-name"
          type="text"
          value={firstName}
          onChange={(e) => setFirstName(e.target.value)}
          required
          disabled={submitting}
        />
        <Input
          label="Password (min 6 characters)"
          id="child-password"
          type="text"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          minLength={6}
          required
          disabled={submitting}
        />

        <AvatarPicker selected={avatar} onSelect={setAvatar} />

        {/* Optional setup fields */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <div className="space-y-1.5">
            <label htmlFor="initial-deposit" className="block text-sm font-semibold text-bark-light">
              Initial Deposit
            </label>
            <div className="flex items-center rounded-xl border border-sand bg-white overflow-hidden focus-within:ring-2 focus-within:ring-forest/30 focus-within:border-forest transition-all">
              <span className="px-3 py-3 bg-cream-dark text-bark-light text-base font-medium border-r border-sand">$</span>
              <input
                type="number"
                id="initial-deposit"
                value={initialDeposit}
                onChange={(e) => setInitialDeposit(e.target.value)}
                placeholder="0.00"
                step="0.01"
                min="0"
                max="999999.99"
                disabled={submitting}
                className="flex-1 min-h-[44px] px-3 py-2 bg-transparent text-bark text-base placeholder:text-bark-light/50 focus:outline-none disabled:cursor-not-allowed"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <label htmlFor="weekly-allowance" className="block text-sm font-semibold text-bark-light">
              Weekly Allowance
            </label>
            <div className="flex items-center rounded-xl border border-sand bg-white overflow-hidden focus-within:ring-2 focus-within:ring-forest/30 focus-within:border-forest transition-all">
              <span className="px-3 py-3 bg-cream-dark text-bark-light text-base font-medium border-r border-sand">$</span>
              <input
                type="number"
                id="weekly-allowance"
                value={weeklyAllowance}
                onChange={(e) => setWeeklyAllowance(e.target.value)}
                placeholder="0.00"
                step="0.01"
                min="0"
                max="999999.99"
                disabled={submitting}
                className="flex-1 min-h-[44px] px-3 py-2 bg-transparent text-bark text-base placeholder:text-bark-light/50 focus:outline-none disabled:cursor-not-allowed"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <label htmlFor="annual-interest" className="block text-sm font-semibold text-bark-light">
              Annual Interest
            </label>
            <div className="flex items-center rounded-xl border border-sand bg-white overflow-hidden focus-within:ring-2 focus-within:ring-forest/30 focus-within:border-forest transition-all">
              <input
                type="number"
                id="annual-interest"
                value={annualInterest}
                onChange={(e) => setAnnualInterest(e.target.value)}
                placeholder="0"
                step="0.01"
                min="0"
                max="100"
                disabled={submitting}
                className="flex-1 min-h-[44px] px-3 py-2 bg-transparent text-bark text-base placeholder:text-bark-light/50 focus:outline-none disabled:cursor-not-allowed"
              />
              <span className="px-3 py-3 bg-cream-dark text-bark-light text-base font-medium border-l border-sand">%</span>
            </div>
          </div>
        </div>

        {error && (
          <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
            <p className="text-sm text-terracotta font-medium">{error}</p>
          </div>
        )}

        <div className="flex gap-3">
          <Button type="submit" loading={submitting} className="flex-1">
            {submitting ? "Creating..." : "Add Child"}
          </Button>
          {onCancel && (
            <Button type="button" variant="secondary" onClick={onCancel} disabled={submitting}>
              Cancel
            </Button>
          )}
        </div>
      </form>

      {created && (
        <div className="mt-4 bg-forest/5 border border-forest/15 rounded-xl p-4">
          <div className="flex items-center gap-2 mb-2">
            <CheckCircle className="h-5 w-5 text-forest" aria-hidden="true" />
            <span className="font-bold text-forest">Account created for {created.first_name}!</span>
          </div>
          <div className="space-y-1 text-sm text-bark-light">
            <p>Login URL: <strong className="text-bark">{window.location.host}{created.login_url}</strong></p>
            <p>Name: <strong className="text-bark">{created.first_name}</strong></p>
            <p>Password: <strong className="text-bark">(the password you just set)</strong></p>
            {setupSummary?.depositAmount && (
              <p>Initial balance: <strong className="text-bark">${setupSummary.depositAmount.toFixed(2)}</strong></p>
            )}
            {setupSummary?.allowanceSet && (
              <p>Weekly allowance: <strong className="text-forest">Active</strong></p>
            )}
            {setupSummary?.interestSet && (
              <p>Annual interest: <strong className="text-forest">Active</strong></p>
            )}
          </div>
          <p className="mt-2 text-xs text-bark-light">Share these credentials with your child.</p>
        </div>
      )}

      {setupWarning && (
        <div className="mt-3 bg-amber/10 border border-amber/20 rounded-xl p-3">
          <p className="text-sm text-amber font-medium">{setupWarning}</p>
        </div>
      )}
    </div>
  );
}
