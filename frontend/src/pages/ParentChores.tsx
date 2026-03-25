import { useState, useEffect, useCallback } from "react";
import { get } from "../api";
import { getChores, getCompletedChores, createChore, updateChore, deleteChore, activateChore, deactivateChore, CreateChoreRequest } from "../api";
import { Chore, ChoreInstance, ChildListResponse } from "../types";
import { useParentUser } from "../hooks/useAuthOutletContext";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Modal from "../components/ui/Modal";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import ChoreForm from "../components/ChoreForm";
import ChoreApprovalQueue from "../components/ChoreApprovalQueue";
import { Plus, Clock, CheckCircle, Users, Pencil, Trash2 } from "lucide-react";

const recurrenceLabels: Record<string, string> = {
  one_time: "One-time",
  daily: "Daily",
  weekly: "Weekly",
  monthly: "Monthly",
};

export default function ParentChoresPage() {
  useParentUser();
  const [chores, setChores] = useState<Chore[]>([]);
  const [children, setChildren] = useState<{ id: number; first_name: string }[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingChore, setEditingChore] = useState<Chore | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [togglingId, setTogglingId] = useState<number | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null);
  const [completedInstances, setCompletedInstances] = useState<ChoreInstance[]>([]);
  const [completedTotal, setCompletedTotal] = useState(0);
  const [loadingCompleted, setLoadingCompleted] = useState(false);

  const PAGE_SIZE = 10;

  const fetchCompleted = useCallback(async (offset: number, append: boolean) => {
    setLoadingCompleted(true);
    try {
      const res = await getCompletedChores(PAGE_SIZE, offset);
      setCompletedInstances((prev) => append ? [...prev, ...(res.instances || [])] : (res.instances || []));
      setCompletedTotal(res.total);
    } catch {
      // Error fetching completed chores
    } finally {
      setLoadingCompleted(false);
    }
  }, []);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [choreRes, childRes] = await Promise.all([
        getChores(),
        get<ChildListResponse>("/children"),
      ]);
      setChores(choreRes.chores || []);
      setChildren(
        (childRes.children || []).map((c) => ({ id: c.id, first_name: c.first_name }))
      );
    } catch {
      // Error fetching data
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
    fetchCompleted(0, false);
  }, [fetchData, fetchCompleted]);

  const handleCreate = async (data: CreateChoreRequest) => {
    setSubmitting(true);
    try {
      await createChore(data);
      setShowForm(false);
      await fetchData();
    } catch {
      // Error creating chore
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdate = async (data: CreateChoreRequest) => {
    if (!editingChore) return;
    setSubmitting(true);
    try {
      await updateChore(editingChore.id, data);
      setEditingChore(null);
      await fetchData();
    } catch {
      // Error updating chore
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (choreId: number) => {
    setDeletingId(choreId);
    try {
      await deleteChore(choreId);
      setConfirmDeleteId(null);
      await fetchData();
    } catch {
      // Error deleting chore
    } finally {
      setDeletingId(null);
    }
  };

  const handleToggleActive = async (choreId: number, isActive: boolean) => {
    setTogglingId(choreId);
    try {
      if (isActive) {
        await deactivateChore(choreId);
      } else {
        await activateChore(choreId);
      }
      await fetchData();
    } catch {
      // Error toggling
    } finally {
      setTogglingId(null);
    }
  };

  if (loading) {
    return <LoadingSpinner message="Loading chores..." />;
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-bark">Chores</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="h-4 w-4" aria-hidden="true" />
          Add Chore
        </Button>
      </div>

      <ChoreApprovalQueue onAction={() => { fetchData(); fetchCompleted(0, false); }} />

      {chores.length === 0 ? (
        <Card>
          <p className="text-center text-bark-light py-8">
            No chores yet. Create one to get started!
          </p>
        </Card>
      ) : (
        <div className="space-y-3">
          {chores.map((chore) => (
            <Card key={chore.id} padding="md">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <h3 className="font-bold text-bark">{chore.name}</h3>
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-cream-dark text-bark-light">
                      <Clock className="h-3 w-3" aria-hidden="true" />
                      {recurrenceLabels[chore.recurrence] || chore.recurrence}
                    </span>
                    {chore.is_active ? (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-forest/10 text-forest">
                        <CheckCircle className="h-3 w-3" aria-hidden="true" />
                        Active
                      </span>
                    ) : (
                      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-sand text-bark-light">
                        Inactive
                      </span>
                    )}
                    {(chore.pending_count ?? 0) > 0 && (
                      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-honey text-bark">
                        {chore.pending_count} pending
                      </span>
                    )}
                  </div>

                  <p className="text-lg font-semibold text-forest mt-1">
                    ${(chore.reward_cents / 100).toFixed(2)}
                  </p>

                  {chore.assignments && chore.assignments.length > 0 && (
                    <div className="flex items-center gap-1 mt-2 text-sm text-bark-light">
                      <Users className="h-3.5 w-3.5" aria-hidden="true" />
                      <span>{chore.assignments.map((a) => a.child_name).join(", ")}</span>
                    </div>
                  )}
                </div>
                <div className="flex items-center gap-2 flex-shrink-0">
                  <button
                    onClick={() => setEditingChore(chore)}
                    className="p-1.5 rounded-lg text-bark-light hover:text-bark hover:bg-sand transition-colors"
                    title="Edit chore"
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                  {confirmDeleteId === chore.id ? (
                    <div className="flex items-center gap-1">
                      <Button
                        variant="primary"
                        onClick={() => handleDelete(chore.id)}
                        loading={deletingId === chore.id}
                        className="!text-xs !min-h-[28px] !px-2 !py-0.5 !bg-terracotta hover:!bg-terracotta/80"
                      >
                        Confirm
                      </Button>
                      <button
                        onClick={() => setConfirmDeleteId(null)}
                        className="text-xs text-bark-light hover:text-bark px-1"
                      >
                        Cancel
                      </button>
                    </div>
                  ) : (
                    <button
                      onClick={() => setConfirmDeleteId(chore.id)}
                      className="p-1.5 rounded-lg text-bark-light hover:text-terracotta hover:bg-terracotta/10 transition-colors"
                      title="Delete chore"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  )}
                  {chore.recurrence !== "one_time" && (
                    <Button
                      variant={chore.is_active ? "secondary" : "primary"}
                      onClick={() => handleToggleActive(chore.id, chore.is_active)}
                      disabled={togglingId === chore.id}
                      loading={togglingId === chore.id}
                      className="text-sm !min-h-[36px] !px-3 !py-1"
                    >
                      {chore.is_active ? "Pause" : "Resume"}
                    </Button>
                  )}
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      {completedInstances.length > 0 && (
        <div className="space-y-3">
          <h2 className="text-lg font-semibold text-bark">Completed</h2>
          {completedInstances.map((instance) => (
            <Card key={instance.id} padding="md">
              <div className="flex items-center justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="font-semibold text-bark">{instance.child_name}</span>
                    <span className="text-bark-light">&middot;</span>
                    <span className="text-bark">{instance.chore_name}</span>
                  </div>
                  {instance.reviewed_at && (
                    <p className="text-xs text-bark-light mt-1">
                      {new Date(instance.reviewed_at).toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" })}
                    </p>
                  )}
                </div>
                <p className="text-lg font-semibold text-forest flex-shrink-0">
                  ${(instance.reward_cents / 100).toFixed(2)}
                </p>
              </div>
            </Card>
          ))}
          {completedInstances.length < completedTotal && (
            <div className="flex justify-center">
              <Button
                variant="secondary"
                onClick={() => fetchCompleted(completedInstances.length, true)}
                loading={loadingCompleted}
              >
                Load More
              </Button>
            </div>
          )}
        </div>
      )}

      <Modal open={showForm} onClose={() => setShowForm(false)} maxWidth="max-w-lg">
        <ChoreForm
          children={children}
          onSubmit={handleCreate}
          onCancel={() => setShowForm(false)}
          loading={submitting}
        />
      </Modal>

      <Modal open={!!editingChore} onClose={() => setEditingChore(null)} maxWidth="max-w-lg">
        {editingChore && (
          <ChoreForm
            children={children}
            onSubmit={handleUpdate}
            onCancel={() => setEditingChore(null)}
            loading={submitting}
            editChore={editingChore}
          />
        )}
      </Modal>
    </div>
  );
}
