import { WithdrawalRequest, WithdrawalRequestStatus } from "../types";
import { useTimezone } from "../context/TimezoneContext";
import { Clock, CheckCircle2, XCircle, Ban } from "lucide-react";

interface WithdrawalRequestCardProps {
  request: WithdrawalRequest;
  onCancel?: (id: number) => void;
  onApprove?: (id: number) => void;
  onDeny?: (id: number) => void;
  loading?: boolean;
}

const statusConfig: Record<
  WithdrawalRequestStatus,
  { label: string; color: string; bgColor: string; icon: typeof Clock }
> = {
  pending: {
    label: "Pending",
    color: "text-honey-dark",
    bgColor: "bg-honey/10",
    icon: Clock,
  },
  approved: {
    label: "Approved",
    color: "text-forest",
    bgColor: "bg-forest/10",
    icon: CheckCircle2,
  },
  denied: {
    label: "Denied",
    color: "text-terracotta",
    bgColor: "bg-terracotta/10",
    icon: XCircle,
  },
  cancelled: {
    label: "Cancelled",
    color: "text-bark-light",
    bgColor: "bg-bark-light/10",
    icon: Ban,
  },
};

export default function WithdrawalRequestCard({
  request,
  onCancel,
  onApprove,
  onDeny,
  loading,
}: WithdrawalRequestCardProps) {
  const timezone = useTimezone();
  const config = statusConfig[request.status];
  const Icon = config.icon;
  const dollars = (request.amount_cents / 100).toFixed(2);

  const formattedDate = new Date(request.created_at).toLocaleDateString(
    undefined,
    {
      year: "numeric",
      month: "short",
      day: "numeric",
      timeZone: timezone,
    }
  );

  return (
    <div className="p-3 border border-sand rounded-xl bg-white">
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-base font-bold text-bark">${dollars}</span>
            <span
              className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${config.color} ${config.bgColor}`}
            >
              <Icon className="h-3 w-3" />
              {config.label}
            </span>
          </div>
          <p className="text-sm text-bark truncate">{request.reason}</p>
          {request.child_name && (
            <p className="text-xs text-bark-light mt-0.5">
              From {request.child_name}
            </p>
          )}
          <p className="text-xs text-bark-light mt-0.5">{formattedDate}</p>
          {request.denial_reason && (
            <p className="text-xs text-terracotta mt-1 italic">
              &ldquo;{request.denial_reason}&rdquo;
            </p>
          )}
        </div>

        <div className="flex items-center gap-2 flex-shrink-0">
          {request.status === "pending" && onApprove && onDeny && (
            <>
              <button
                onClick={() => onApprove(request.id)}
                disabled={loading}
                className="px-3 py-1.5 text-xs font-semibold text-white bg-forest rounded-lg hover:bg-forest/90 disabled:opacity-50 transition-colors"
              >
                Approve
              </button>
              <button
                onClick={() => onDeny(request.id)}
                disabled={loading}
                className="px-3 py-1.5 text-xs font-semibold text-terracotta bg-terracotta/10 rounded-lg hover:bg-terracotta/20 disabled:opacity-50 transition-colors"
              >
                Deny
              </button>
            </>
          )}
          {request.status === "pending" && onCancel && (
            <button
              onClick={() => onCancel(request.id)}
              disabled={loading}
              className="px-3 py-1.5 text-xs font-semibold text-bark-light bg-bark-light/10 rounded-lg hover:bg-bark-light/20 disabled:opacity-50 transition-colors"
            >
              Cancel
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
