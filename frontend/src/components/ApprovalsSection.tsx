import { useState, useEffect, useCallback } from "react";
import { getPendingChores, getWithdrawalRequests } from "../api";
import { WithdrawalRequest } from "../types";
import ChoreApprovalQueue from "./ChoreApprovalQueue";
import PendingWithdrawalRequests from "./PendingWithdrawalRequests";

interface ApprovalsSectionProps {
  onUpdated: () => void;
}

export default function ApprovalsSection({ onUpdated }: ApprovalsSectionProps) {
  const [pendingChoreCount, setPendingChoreCount] = useState(0);
  const [withdrawalRequests, setWithdrawalRequests] = useState<WithdrawalRequest[]>([]);
  const [loaded, setLoaded] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      const [choreRes, wrRes] = await Promise.all([
        getPendingChores(),
        getWithdrawalRequests({ status: "pending" }),
      ]);
      setPendingChoreCount((choreRes.instances || []).length);
      setWithdrawalRequests(wrRes.withdrawal_requests || []);
    } catch {
      // Silently fail
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleChoreAction = () => {
    fetchData();
    onUpdated();
  };

  const handleWithdrawalUpdated = () => {
    fetchData();
    onUpdated();
  };

  if (!loaded) return null;

  const pendingWithdrawals = withdrawalRequests.filter((r) => r.status === "pending");
  if (pendingChoreCount === 0 && pendingWithdrawals.length === 0) return null;

  const totalCount = pendingChoreCount + pendingWithdrawals.length;

  return (
    <section className="space-y-3">
      <h2 className="text-lg font-semibold text-bark">
        Approvals{" "}
        <span className="text-sm font-normal text-bark-light">
          ({totalCount})
        </span>
      </h2>
      <ChoreApprovalQueue onAction={handleChoreAction} />
      <PendingWithdrawalRequests
        requests={withdrawalRequests}
        onUpdated={handleWithdrawalUpdated}
      />
    </section>
  );
}
