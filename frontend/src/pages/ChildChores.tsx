import { useEffect, useState, useCallback } from "react";
import { getChildChores, completeChore, getChoreEarnings } from "../api";
import { ChoreInstance, ChoreEarningsResponse } from "../types";
import ChoreCard from "../components/ChoreCard";
import Card from "../components/ui/Card";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import { useChildUser } from "../hooks/useAuthOutletContext";
import { DollarSign, Trophy } from "lucide-react";

export default function ChildChoresPage() {
  useChildUser();
  const [available, setAvailable] = useState<ChoreInstance[]>([]);
  const [pending, setPending] = useState<ChoreInstance[]>([]);
  const [completed, setCompleted] = useState<ChoreInstance[]>([]);
  const [loading, setLoading] = useState(true);
  const [completingId, setCompletingId] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [earnings, setEarnings] = useState<ChoreEarningsResponse | null>(null);

  const fetchChores = useCallback(async () => {
    try {
      const [choreRes, earningsRes] = await Promise.all([
        getChildChores(),
        getChoreEarnings().catch(() => null),
      ]);
      setAvailable(choreRes.available || []);
      setPending(choreRes.pending || []);
      setCompleted((choreRes.completed || []).slice(0, 10));
      setEarnings(earningsRes);
      setError(null);
    } catch {
      setError("Failed to load chores. Please try again.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchChores();
  }, [fetchChores]);

  const handleComplete = async (id: number) => {
    setCompletingId(id);
    setError(null);
    try {
      await completeChore(id);
      await fetchChores();
    } catch {
      setError("Failed to complete chore. Please try again.");
    } finally {
      setCompletingId(null);
    }
  };

  if (loading) {
    return <LoadingSpinner message="Loading chores..." />;
  }

  const isEmpty = available.length === 0 && pending.length === 0 && completed.length === 0;

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold text-bark">My Chores</h1>

      {error && (
        <div className="bg-terracotta/10 text-terracotta px-4 py-3 rounded-xl text-sm font-medium">
          {error}
        </div>
      )}

      {earnings && (earnings.total_earned_cents > 0 || earnings.chores_completed > 0) && (
        <Card padding="md">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 flex-1">
              <DollarSign className="h-5 w-5 text-forest" aria-hidden="true" />
              <div>
                <p className="text-sm text-bark-light">Total Earned</p>
                <p className="text-xl font-bold text-forest">
                  ${(earnings.total_earned_cents / 100).toFixed(2)}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2 flex-1">
              <Trophy className="h-5 w-5 text-honey-dark" aria-hidden="true" />
              <div>
                <p className="text-sm text-bark-light">Chores Completed</p>
                <p className="text-xl font-bold text-bark">
                  {earnings.chores_completed}
                </p>
              </div>
            </div>
          </div>
        </Card>
      )}

      {isEmpty ? (
        <div className="text-center py-12 text-bark-light">
          <p className="text-lg font-medium">No chores assigned yet!</p>
          <p className="text-sm mt-1">Check back later for new chores.</p>
        </div>
      ) : (
        <>
          {available.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold text-bark mb-3">Available Chores</h2>
              <div className="space-y-3">
                {available.map((instance) => (
                  <ChoreCard
                    key={instance.id}
                    instance={instance}
                    onComplete={handleComplete}
                    loading={completingId === instance.id}
                  />
                ))}
              </div>
            </section>
          )}

          {pending.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold text-bark mb-3">Pending Approval</h2>
              <div className="space-y-3">
                {pending.map((instance) => (
                  <ChoreCard key={instance.id} instance={instance} />
                ))}
              </div>
            </section>
          )}

          {completed.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold text-bark mb-3">Completed</h2>
              <div className="space-y-3">
                {completed.map((instance) => (
                  <ChoreCard key={instance.id} instance={instance} />
                ))}
              </div>
            </section>
          )}
        </>
      )}
    </div>
  );
}
