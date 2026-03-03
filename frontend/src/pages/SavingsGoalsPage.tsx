import { useEffect, useState, useCallback } from "react";
import { getSavingsGoals, createSavingsGoal } from "../api";
import { SavingsGoal } from "../types";
import { useChildUser } from "../hooks/useAuthOutletContext";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import GoalCard from "../components/GoalCard";
import GoalForm from "../components/GoalForm";
import { Plus } from "lucide-react";

export default function SavingsGoalsPage() {
  const user = useChildUser();
  const [goals, setGoals] = useState<SavingsGoal[]>([]);
  const [availableBalanceCents, setAvailableBalanceCents] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);

  const activeGoals = goals.filter((g) => g.status === "active");
  const completedGoals = goals.filter((g) => g.status === "completed");

  const fetchGoals = useCallback(async () => {
    try {
      const res = await getSavingsGoals(user.user_id);
      setGoals(res.goals);
      setAvailableBalanceCents(res.available_balance_cents);
    } catch {
      // Silently fail
    } finally {
      setLoading(false);
    }
  }, [user.user_id]);

  useEffect(() => {
    fetchGoals();
  }, [fetchGoals]);

  const handleCreate = async (data: { name: string; target_cents: number; emoji?: string; target_date?: string }) => {
    await createSavingsGoal(user.user_id, data);
    setShowForm(false);
    await fetchGoals();
  };

  if (loading) {
    return (
      <div className="max-w-[480px] mx-auto py-12">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="max-w-[480px] mx-auto space-y-6 animate-fade-in-up">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-forest">My Savings Goals</h2>
        {!showForm && (
          <Button
            variant="primary"
            onClick={() => setShowForm(true)}
            disabled={activeGoals.length >= 5}
            className="!min-h-[40px] !px-4 !py-2 text-sm"
            title={activeGoals.length >= 5 ? "Maximum of 5 active goals" : undefined}
          >
            <Plus className="h-4 w-4" />
            Add Goal
          </Button>
        )}
      </div>

      {/* Available balance */}
      <p className="text-sm text-bark-light">
        Available balance: <span className="font-semibold text-bark">${(availableBalanceCents / 100).toFixed(2)}</span>
      </p>

      {/* Create form */}
      {showForm && (
        <Card padding="md">
          <GoalForm onSubmit={handleCreate} onCancel={() => setShowForm(false)} />
        </Card>
      )}

      {/* Active goals */}
      {activeGoals.length > 0 ? (
        <div className="space-y-3">
          {activeGoals.map((goal) => (
            <GoalCard key={goal.id} goal={goal} />
          ))}
        </div>
      ) : (
        !showForm && (
          <Card padding="md" className="text-center">
            <p className="text-bark-light">No savings goals yet. Create one to start saving!</p>
          </Card>
        )
      )}

      {/* Completed goals */}
      {completedGoals.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-bark-light">Completed Goals</h3>
          {completedGoals.map((goal) => (
            <div key={goal.id} className="opacity-70">
              <GoalCard goal={goal} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
