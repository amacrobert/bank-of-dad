import { useEffect, useState, useCallback } from "react";
import { getSavingsGoals, createSavingsGoal, allocateToGoal, updateSavingsGoal, deleteSavingsGoal } from "../api";
import { SavingsGoal, AllocateResponse } from "../types";
import { useChildUser } from "../hooks/useAuthOutletContext";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import GoalCard from "../components/GoalCard";
import GoalForm from "../components/GoalForm";
import ConfettiCelebration from "../components/ConfettiCelebration";
import { Plus, PencilLine, Trash2 } from "lucide-react";

export default function SavingsGoalsPage() {
  const user = useChildUser();
  const [goals, setGoals] = useState<SavingsGoal[]>([]);
  const [availableBalanceCents, setAvailableBalanceCents] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [showConfetti, setShowConfetti] = useState(false);
  const [completedGoalName, setCompletedGoalName] = useState("");
  const [editingGoal, setEditingGoal] = useState<SavingsGoal | null>(null);
  const [deletingGoal, setDeletingGoal] = useState<SavingsGoal | null>(null);
  const [deleting, setDeleting] = useState(false);

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

  const handleUpdate = async (data: { name: string; target_cents: number; emoji?: string; target_date?: string }) => {
    if (!editingGoal) return;
    await updateSavingsGoal(user.user_id, editingGoal.id, data);
    setEditingGoal(null);
    await fetchGoals();
  };

  const handleAllocate = async (goalId: number, amountCents: number) => {
    const res: AllocateResponse = await allocateToGoal(user.user_id, goalId, amountCents);
    if (res.completed) {
      setCompletedGoalName(res.goal.name);
      setShowConfetti(true);
    }
    await fetchGoals();
  };

  const handleDelete = async () => {
    if (!deletingGoal) return;

    setDeleting(true);
    try {
      await deleteSavingsGoal(user.user_id, deletingGoal.id);
      setDeletingGoal(null);
      if (editingGoal?.id === deletingGoal.id) {
        setEditingGoal(null);
      }
      await fetchGoals();
    } finally {
      setDeleting(false);
    }
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
      <ConfettiCelebration show={showConfetti} onComplete={() => setShowConfetti(false)} />

      {/* Goal achieved message */}
      {completedGoalName && showConfetti && (
        <Card padding="md" className="text-center bg-sage-light/30 border-forest/20 animate-scale-in">
          <p className="text-lg font-bold text-forest">Goal Achieved!</p>
          <p className="text-sm text-bark-light">You completed "{completedGoalName}"</p>
        </Card>
      )}

      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-forest">My Savings Goals</h2>
        {!showForm && (
          <Button
            variant="primary"
            onClick={() => {
              setEditingGoal(null);
              setShowForm(true);
            }}
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

      {editingGoal && (
        <Card padding="md" className="space-y-4 border-forest/20 bg-sage-light/20">
          <div className="flex items-center gap-2">
            <PencilLine className="h-4 w-4 text-forest" />
            <p className="text-sm font-semibold text-forest">Edit Goal</p>
          </div>
          <GoalForm
            initialGoal={editingGoal}
            onSubmit={handleUpdate}
            onCancel={() => setEditingGoal(null)}
          />
        </Card>
      )}

      {deletingGoal && (
        <Card padding="md" className="space-y-4 border-terracotta/20 bg-terracotta/5">
          <div className="flex items-center gap-2">
            <Trash2 className="h-4 w-4 text-terracotta" />
            <p className="text-sm font-semibold text-terracotta">Delete Goal</p>
          </div>
          <div className="space-y-1 text-sm text-bark-light">
            <p>Delete "{deletingGoal.name}"?</p>
            <p>${(deletingGoal.saved_cents / 100).toFixed(2)} will return to your available balance.</p>
          </div>
          <div className="flex gap-3">
            <Button variant="danger" loading={deleting} onClick={handleDelete} className="flex-1">
              Delete Goal
            </Button>
            <Button
              type="button"
              variant="secondary"
              onClick={() => setDeletingGoal(null)}
              disabled={deleting}
            >
              Cancel
            </Button>
          </div>
        </Card>
      )}

      {/* Active goals */}
      {activeGoals.length > 0 ? (
        <div className="space-y-3">
          {activeGoals.map((goal) => (
            <GoalCard
              key={goal.id}
              goal={goal}
              childId={user.user_id}
              onAllocate={handleAllocate}
              onEdit={(selectedGoal) => {
                setShowForm(false);
                setDeletingGoal(null);
                setEditingGoal(selectedGoal);
              }}
              onDelete={(selectedGoal) => {
                setShowForm(false);
                setEditingGoal(null);
                setDeletingGoal(selectedGoal);
              }}
            />
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
              <GoalCard goal={goal} childId={user.user_id} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
