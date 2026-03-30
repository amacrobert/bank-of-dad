import { useState, useEffect, useCallback } from "react";
import {
  getPendingChores,
  getWithdrawalRequests,
  approveChore,
  rejectChore,
  approveWithdrawalRequest,
  denyWithdrawalRequest,
  ApiRequestError,
} from "../api";
import { ChoreInstance, WithdrawalRequest } from "../types";
import Card from "./ui/Card";
import Button from "./ui/Button";
import Input from "./ui/Input";
import Modal from "./ui/Modal";
import { CheckCircle, X } from "lucide-react";

interface ApprovalsSectionProps {
  onUpdated: () => void;
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

export default function ApprovalsSection({ onUpdated }: ApprovalsSectionProps) {
  const [choreInstances, setChoreInstances] = useState<ChoreInstance[]>([]);
  const [withdrawalRequests, setWithdrawalRequests] = useState<WithdrawalRequest[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [actionId, setActionId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Chore rejection
  const [rejectingChoreId, setRejectingChoreId] = useState<number | null>(null);
  const [choreRejectReason, setChoreRejectReason] = useState("");

  // Withdrawal denial
  const [showDenyModal, setShowDenyModal] = useState(false);
  const [denyingRequestId, setDenyingRequestId] = useState<number | null>(null);
  const [denyReason, setDenyReason] = useState("");

  // Goal impact warning
  const [goalWarning, setGoalWarning] = useState<{
    requestId: number;
    message: string;
  } | null>(null);

  const fetchData = useCallback(async () => {
    try {
      const [choreRes, wrRes] = await Promise.all([
        getPendingChores(),
        getWithdrawalRequests({ status: "pending" }),
      ]);
      setChoreInstances(choreRes.instances || []);
      setWithdrawalRequests(
        (wrRes.withdrawal_requests || []).filter(
          (r: WithdrawalRequest) => r.status === "pending"
        )
      );
    } catch {
      // Silently fail
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const refresh = () => {
    fetchData();
    onUpdated();
  };

  // --- Chore actions ---

  const handleApproveChore = async (id: number) => {
    setActionId(`chore-${id}`);
    setError(null);
    try {
      await approveChore(id);
      refresh();
    } catch {
      setError("Failed to approve chore.");
    } finally {
      setActionId(null);
    }
  };

  const handleRejectChoreConfirm = async (id: number) => {
    setActionId(`chore-${id}`);
    setError(null);
    try {
      await rejectChore(id, choreRejectReason || undefined);
      setRejectingChoreId(null);
      setChoreRejectReason("");
      refresh();
    } catch {
      setError("Failed to reject chore.");
    } finally {
      setActionId(null);
    }
  };

  // --- Withdrawal actions ---

  const handleApproveWithdrawal = async (id: number, confirmGoalImpact = false) => {
    setActionId(`wr-${id}`);
    setError(null);
    try {
      await approveWithdrawalRequest(id, { confirm_goal_impact: confirmGoalImpact });
      setGoalWarning(null);
      refresh();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        if (err.body.error === "goal_impact_warning") {
          setGoalWarning({
            requestId: id,
            message: err.body.message || "This approval will affect savings goals.",
          });
        } else {
          setError(err.body.message || err.body.error);
        }
      } else {
        setError("Failed to approve request.");
      }
    } finally {
      setActionId(null);
    }
  };

  const handleDenyClick = (id: number) => {
    setDenyingRequestId(id);
    setDenyReason("");
    setShowDenyModal(true);
  };

  const handleDenySubmit = async () => {
    if (denyingRequestId === null) return;
    setActionId(`wr-${denyingRequestId}`);
    setError(null);
    try {
      await denyWithdrawalRequest(denyingRequestId, {
        reason: denyReason.trim() || undefined,
      });
      setShowDenyModal(false);
      setDenyingRequestId(null);
      refresh();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to deny request.");
      }
    } finally {
      setActionId(null);
    }
  };

  if (!loaded) return null;
  if (choreInstances.length === 0 && withdrawalRequests.length === 0) return null;

  const totalCount = choreInstances.length + withdrawalRequests.length;

  return (
    <section className="space-y-3 mb-6">
      <h2 className="text-lg font-semibold text-bark">
        Approvals{" "}
        <span className="text-sm font-normal text-bark-light">
          ({totalCount})
        </span>
      </h2>

      {error && (
        <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
          <p className="text-sm text-terracotta font-medium">{error}</p>
        </div>
      )}

      {goalWarning && (
        <div className="bg-honey/10 border border-honey/30 rounded-xl p-3 space-y-2">
          <p className="text-sm text-honey-dark font-medium">
            {goalWarning.message}
          </p>
          <div className="flex gap-2">
            <Button
              variant="primary"
              onClick={() => handleApproveWithdrawal(goalWarning.requestId, true)}
              loading={actionId === `wr-${goalWarning.requestId}`}
              className="text-sm !min-h-[36px] !px-3 !py-1"
            >
              Approve Anyway
            </Button>
            <Button
              variant="ghost"
              onClick={() => setGoalWarning(null)}
              className="text-sm !min-h-[36px] !px-3 !py-1"
            >
              Cancel
            </Button>
          </div>
        </div>
      )}

      {/* Chores */}
      {choreInstances.length > 0 && (
        <div className="space-y-2">
          <h3 className="text-sm font-semibold text-bark-light uppercase tracking-wide">
            Chores
          </h3>
          {choreInstances.map((inst) => (
            <Card key={`chore-${inst.id}`} padding="md">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm font-medium text-bark-light">
                      {inst.child_name}
                    </span>
                    <span className="text-bark-light">&middot;</span>
                    <h4 className="font-bold text-bark">{inst.chore_name}</h4>
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

                {rejectingChoreId === inst.id ? (
                  <div className="flex flex-col gap-2 flex-shrink-0">
                    <input
                      type="text"
                      placeholder="Reason (optional)"
                      value={choreRejectReason}
                      onChange={(e) => setChoreRejectReason(e.target.value)}
                      className="border border-sand rounded-lg px-3 py-2 text-sm w-48 focus:outline-none focus:ring-2 focus:ring-forest/40"
                    />
                    <div className="flex gap-2">
                      <Button
                        variant="danger"
                        onClick={() => handleRejectChoreConfirm(inst.id)}
                        loading={actionId === `chore-${inst.id}`}
                        className="text-sm !min-h-[36px] !px-3 !py-1"
                      >
                        Confirm
                      </Button>
                      <Button
                        variant="ghost"
                        onClick={() => {
                          setRejectingChoreId(null);
                          setChoreRejectReason("");
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
                      onClick={() => handleApproveChore(inst.id)}
                      loading={actionId === `chore-${inst.id}`}
                      className="text-sm !min-h-[36px] !px-3 !py-1"
                    >
                      <CheckCircle className="h-4 w-4" aria-hidden="true" />
                      Approve
                    </Button>
                    <Button
                      variant="ghost"
                      onClick={() => setRejectingChoreId(inst.id)}
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
        </div>
      )}

      {/* Withdrawal Requests */}
      {withdrawalRequests.length > 0 && (
        <div className="space-y-2">
          <h3 className="text-sm font-semibold text-bark-light uppercase tracking-wide">
            Withdrawal Requests
          </h3>
          {withdrawalRequests.map((req) => (
            <Card key={`wr-${req.id}`} padding="md">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm font-medium text-bark-light">
                      {req.child_name}
                    </span>
                    <span className="text-bark-light">&middot;</span>
                    <h4 className="font-bold text-bark">{req.reason}</h4>
                  </div>
                  <p className="text-lg font-semibold text-forest mt-1">
                    ${(req.amount_cents / 100).toFixed(2)}
                  </p>
                  <p className="text-xs text-bark-light mt-1">
                    Requested {formatDateTime(req.created_at)}
                  </p>
                </div>

                <div className="flex gap-2 flex-shrink-0">
                  <Button
                    onClick={() => handleApproveWithdrawal(req.id)}
                    loading={actionId === `wr-${req.id}`}
                    className="text-sm !min-h-[36px] !px-3 !py-1"
                  >
                    <CheckCircle className="h-4 w-4" aria-hidden="true" />
                    Approve
                  </Button>
                  <Button
                    variant="ghost"
                    onClick={() => handleDenyClick(req.id)}
                    className="text-sm !min-h-[36px] !px-3 !py-1 !text-terracotta hover:!bg-terracotta/10"
                  >
                    <X className="h-4 w-4" aria-hidden="true" />
                    Deny
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      <Modal open={showDenyModal} onClose={() => setShowDenyModal(false)}>
        <div>
          <h4 className="text-base font-bold text-bark mb-3">
            Deny Withdrawal Request
          </h4>
          <Input
            label="Reason (optional)"
            id="deny-reason"
            type="text"
            value={denyReason}
            onChange={(e) => setDenyReason(e.target.value)}
            placeholder="e.g., Save up a bit more first"
            maxLength={500}
            disabled={actionId !== null}
          />
          <div className="flex gap-3 mt-4">
            <Button
              variant="danger"
              onClick={handleDenySubmit}
              loading={actionId !== null}
              className="flex-1"
            >
              {actionId !== null ? "Denying..." : "Deny Request"}
            </Button>
            <Button
              variant="secondary"
              onClick={() => setShowDenyModal(false)}
              disabled={actionId !== null}
            >
              Cancel
            </Button>
          </div>
        </div>
      </Modal>
    </section>
  );
}
