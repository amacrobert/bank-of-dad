import { useState } from "react";
import {
  approveWithdrawalRequest,
  denyWithdrawalRequest,
  ApiRequestError,
} from "../api";
import { WithdrawalRequest } from "../types";
import WithdrawalRequestCard from "./WithdrawalRequestCard";
import Modal from "./ui/Modal";
import Button from "./ui/Button";
import Input from "./ui/Input";

interface PendingWithdrawalRequestsProps {
  requests: WithdrawalRequest[];
  onUpdated: () => void;
}

export default function PendingWithdrawalRequests({
  requests,
  onUpdated,
}: PendingWithdrawalRequestsProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showDenyModal, setShowDenyModal] = useState(false);
  const [denyingRequestId, setDenyingRequestId] = useState<number | null>(null);
  const [denyReason, setDenyReason] = useState("");
  const [goalWarning, setGoalWarning] = useState<{
    requestId: number;
    message: string;
  } | null>(null);

  const pendingRequests = requests.filter((r) => r.status === "pending");

  if (pendingRequests.length === 0) return null;

  const handleApprove = async (id: number, confirmGoalImpact = false) => {
    setLoading(true);
    setError(null);
    try {
      await approveWithdrawalRequest(id, {
        confirm_goal_impact: confirmGoalImpact,
      });
      setGoalWarning(null);
      onUpdated();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        if (err.body.error === "goal_impact_warning") {
          setGoalWarning({ requestId: id, message: err.body.message || "This approval will affect savings goals." });
        } else {
          setError(err.body.message || err.body.error);
        }
      } else {
        setError("Failed to approve request.");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDenyClick = (id: number) => {
    setDenyingRequestId(id);
    setDenyReason("");
    setShowDenyModal(true);
  };

  const handleDenySubmit = async () => {
    if (denyingRequestId === null) return;
    setLoading(true);
    setError(null);
    try {
      await denyWithdrawalRequest(denyingRequestId, {
        reason: denyReason.trim() || undefined,
      });
      setShowDenyModal(false);
      setDenyingRequestId(null);
      onUpdated();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to deny request.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-2">
      <h4 className="text-sm font-semibold text-bark-light uppercase tracking-wide">
        Pending Withdrawal Requests
      </h4>

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
              onClick={() =>
                handleApprove(goalWarning.requestId, true)
              }
              loading={loading}
              className="text-xs"
            >
              Approve Anyway
            </Button>
            <Button
              variant="secondary"
              onClick={() => setGoalWarning(null)}
              className="text-xs"
            >
              Cancel
            </Button>
          </div>
        </div>
      )}

      <div className="space-y-2">
        {pendingRequests.map((req) => (
          <WithdrawalRequestCard
            key={req.id}
            request={req}
            onApprove={(id) => handleApprove(id)}
            onDeny={handleDenyClick}
            loading={loading}
          />
        ))}
      </div>

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
            disabled={loading}
          />
          <div className="flex gap-3 mt-4">
            <Button
              variant="danger"
              onClick={handleDenySubmit}
              loading={loading}
              className="flex-1"
            >
              {loading ? "Denying..." : "Deny Request"}
            </Button>
            <Button
              variant="secondary"
              onClick={() => setShowDenyModal(false)}
              disabled={loading}
            >
              Cancel
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
