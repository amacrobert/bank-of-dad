import { useEffect, useState, useCallback } from "react";
import { getSavingsGoals, createSavingsGoal, allocateToGoal, updateSavingsGoal, deleteSavingsGoal } from "../api";
import { SavingsGoal, AllocateResponse } from "../types";
import { useChildUser } from "../hooks/useAuthOutletContext";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Modal from "../components/ui/Modal";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import GoalCard from "../components/GoalCard";
import GoalForm from "../components/GoalForm";
import ConfettiCelebration from "../components/ConfettiCelebration";
import { Plus } from "lucide-react";

export default function SavingsGoalsPage() {
  const user = useChildUser();
  const [goals, setGoals] = useState<SavingsGoal[]>([]);
  const [availableBalanceCents, setAvailableBalanceCents] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingGoal, setEditingGoal] = useState<SavingsGoal | null>(null);
  const [deletingGoal, setDeletingGoal] = useState<SavingsGoal | null>(null);
  const [showConfetti, setShowConfetti] = useState(false);
  const [completedGoalName, setCompletedGoalName] = useState("");
  const [showAchievedCard, setShowAchievedCard] = useState(false);

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

  const handleCreate = async (data: { name: string; target_cents: number; emoji?: string }) => {
    await createSavingsGoal(user.user_id, data);
    setShowForm(false);
    await fetchGoals();
  };

  const handleAllocate = async (goalId: number, amountCents: number) => {
    const res: AllocateResponse = await allocateToGoal(user.user_id, goalId, amountCents);
    if (res.completed) {
      setCompletedGoalName(res.goal.name);
      setShowConfetti(true);
      setShowAchievedCard(true);
    }
    await fetchGoals();
  };

  const handleEdit = (goal: SavingsGoal) => {
    setEditingGoal(goal);
    setShowForm(false);
  };

  const handleEditSubmit = async (data: { name: string; target_cents: number; emoji?: string }) => {
    if (!editingGoal) return;
    await updateSavingsGoal(user.user_id, editingGoal.id, data);
    setEditingGoal(null);
    await fetchGoals();
  };

  const handleDelete = (goal: SavingsGoal) => {
    setDeletingGoal(goal);
  };

  const handleDeleteConfirm = async () => {
    if (!deletingGoal) return;
    await deleteSavingsGoal(user.user_id, deletingGoal.id);
    setDeletingGoal(null);
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
      <ConfettiCelebration show={showConfetti} onComplete={() => setShowConfetti(false)} />

      {/* Goal achieved message */}
      {completedGoalName && showAchievedCard && (
        <Card padding="md" className="text-center bg-sage-light/30 border-forest/20 animate-scale-in">
          <p className="text-lg font-bold text-forest">Goal Achieved!</p>
          <p className="text-sm mt-3">
            You completed "{completedGoalName}"!
          </p>
          <p className="text-sm text-bark-light mt-3">
            The funds have been returned to your available balance. If this goal was for a purchase, ask your grown-up to make a withdrawal for you.
          </p>
          <p className="text-sm mt-3">
            Congratulations on saving and hitting your goal!
          </p>
          <Button
            variant="primary"
            onClick={() => { setShowAchievedCard(false); setCompletedGoalName(""); }}
            className="mt-3"
          >
            Ok
          </Button>
        </Card>
      )}

      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-forest">My Savings Goals</h2>
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
      </div>

      {/* Available balance */}
      <p className="text-sm text-bark-light">
        Available balance: <span className="font-semibold text-bark">${(availableBalanceCents / 100).toFixed(2)}</span>
      </p>

      {/* Create form */}
      <Modal open={showForm} onClose={() => setShowForm(false)}>
        <GoalForm onSubmit={handleCreate} onCancel={() => setShowForm(false)} />
      </Modal>

      {/* Edit form */}
      <Modal open={!!editingGoal} onClose={() => setEditingGoal(null)}>
        {editingGoal && (
          <>
            <h3 className="text-sm font-semibold text-bark-light mb-3">Edit Goal</h3>
            <GoalForm
              onSubmit={handleEditSubmit}
              onCancel={() => setEditingGoal(null)}
              initialGoal={editingGoal}
            />
          </>
        )}
      </Modal>

      {/* Delete confirmation modal */}
      <Modal open={!!deletingGoal} onClose={() => setDeletingGoal(null)} maxWidth="max-w-sm">
        {deletingGoal && (
          <div className="border-terracotta/30 bg-terracotta/5 -m-6 p-6 rounded-2xl">
            <p className="text-sm text-bark mb-3">
              Delete "{deletingGoal.name}"? {deletingGoal.saved_cents > 0 && deletingGoal.status === "active" && (
                <span className="font-semibold">
                  ${(deletingGoal.saved_cents / 100).toFixed(2)} will return to your available balance.
                </span>
              )}
            </p>
            <div className="flex gap-3">
              <Button
                variant="primary"
                onClick={handleDeleteConfirm}
                className="!bg-terracotta hover:!bg-terracotta/90 flex-1"
              >
                Delete
              </Button>
              <Button variant="secondary" onClick={() => setDeletingGoal(null)}>
                Cancel
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Active goals */}
      {activeGoals.length > 0 ? (
        <div className="space-y-3">
          {activeGoals.map((goal) => (
            <GoalCard
              key={goal.id}
              goal={goal}
              childId={user.user_id}
              onAllocate={handleAllocate}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          ))}
        </div>
      ) : (
        !showForm && !editingGoal && (
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
              <GoalCard goal={goal} childId={user.user_id} onDelete={handleDelete} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
