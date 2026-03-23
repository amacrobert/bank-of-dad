import { useEffect, useState, useCallback } from "react";
import { getPendingChores, approveChore, rejectChore } from "../api";
import { ChoreInstance } from "../types";
import Card from "./ui/Card";
import Button from "./ui/Button";
import { CheckCircle, X } from "lucide-react";

interface ChoreApprovalQueueProps {
  onAction?: () => void;
}

function formatDateTime(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export default function ChoreApprovalQueue({ onAction }: ChoreApprovalQueueProps) {
  const [instances, setInstances] = useState<ChoreInstance[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionId, setActionId] = useState<number | null>(null);
  const [rejectingId, setRejectingId] = useState<number | null>(null);
  const [rejectReason, setRejectReason] = useState("");

  const fetchPending = useCallback(async () => {
    try {
      const res = await getPendingChores();
      setInstances(res.instances || []);
    } catch {
      // Silently fail; parent page handles main data
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchPending();
  }, [fetchPending]);

  const handleApprove = async (id: number) => {
    setActionId(id);
    try {
      await approveChore(id);
      await fetchPending();
      onAction?.();
    } catch {
      // Error approving
    } finally {
      setActionId(null);
    }
  };

  const handleRejectConfirm = async (id: number) => {
    setActionId(id);
    try {
      await rejectChore(id, rejectReason || undefined);
      setRejectingId(null);
      setRejectReason("");
      await fetchPending();
      onAction?.();
    } catch {
      // Error rejecting
    } finally {
      setActionId(null);
    }
  };

  if (loading) {
    return null;
  }

  if (instances.length === 0) {
    return null;
  }

  return (
    <section className="space-y-3">
      <h2 className="text-lg font-semibold text-bark">Pending Approval</h2>
      {instances.map((inst) => (
        <Card key={inst.id} padding="md">
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-sm font-medium text-bark-light">
                  {inst.child_name}
                </span>
                <span className="text-bark-light">&middot;</span>
                <h3 className="font-bold text-bark">{inst.chore_name}</h3>
              </div>
              <p className="text-lg font-semibold text-forest mt-1">
                ${(inst.reward_cents / 100).toFixed(2)}
              </p>
              {inst.completed_at && (
                <p className="text-xs text-bark-light mt-1">
                  Completed {formatDateTime(inst.completed_at)}
                </p>
              )}
            </div>

            {rejectingId === inst.id ? (
              <div className="flex flex-col gap-2 flex-shrink-0">
                <input
                  type="text"
                  placeholder="Reason (optional)"
                  value={rejectReason}
                  onChange={(e) => setRejectReason(e.target.value)}
                  className="border border-sand rounded-lg px-3 py-2 text-sm w-48 focus:outline-none focus:ring-2 focus:ring-forest/40"
                />
                <div className="flex gap-2">
                  <Button
                    variant="danger"
                    onClick={() => handleRejectConfirm(inst.id)}
                    loading={actionId === inst.id}
                    className="text-sm !min-h-[36px] !px-3 !py-1"
                  >
                    Confirm
                  </Button>
                  <Button
                    variant="ghost"
                    onClick={() => {
                      setRejectingId(null);
                      setRejectReason("");
                    }}
                    className="text-sm !min-h-[36px] !px-3 !py-1"
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            ) : (
              <div className="flex gap-2 flex-shrink-0">
                <Button
                  onClick={() => handleApprove(inst.id)}
                  loading={actionId === inst.id}
                  className="text-sm !min-h-[36px] !px-3 !py-1"
                >
                  <CheckCircle className="h-4 w-4" aria-hidden="true" />
                  Approve
                </Button>
                <Button
                  variant="ghost"
                  onClick={() => setRejectingId(inst.id)}
                  className="text-sm !min-h-[36px] !px-3 !py-1 !text-terracotta hover:!bg-terracotta/10"
                >
                  <X className="h-4 w-4" aria-hidden="true" />
                  Reject
                </Button>
              </div>
            )}
          </div>
        </Card>
      ))}
    </section>
  );
}
