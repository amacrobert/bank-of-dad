import { ApiError } from "./types";
import { getAccessToken, clearTokens, refreshTokens } from "./auth";

const API_BASE = import.meta.env.VITE_API_URL || '';

export class ApiRequestError extends Error {
  status: number;
  body: ApiError;

  constructor(status: number, body: ApiError) {
    super(body.message || body.error);
    this.status = status;
    this.body = body;
  }
}

async function request<T>(
  path: string,
  init?: RequestInit
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  const token = getAccessToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}/api${path}`, {
    headers,
    ...init,
  });

  if (!res.ok) {
    let body: ApiError;
    try {
      body = await res.json();
    } catch {
      body = { error: "Unknown error" };
    }

    if (res.status === 401) {
      // Attempt to refresh tokens before giving up
      const refreshed = await refreshTokens();
      if (refreshed) {
        // Retry the original request with new token
        const retryHeaders: Record<string, string> = {
          "Content-Type": "application/json",
        };
        const newToken = getAccessToken();
        if (newToken) {
          retryHeaders["Authorization"] = `Bearer ${newToken}`;
        }
        const retryRes = await fetch(`${API_BASE}/api${path}`, {
          ...init,
          headers: retryHeaders,
        });
        if (retryRes.ok) {
          if (retryRes.status === 204) {
            return undefined as T;
          }
          return retryRes.json();
        }
      }
      clearTokens();
      window.location.href = "/";
      throw new ApiRequestError(res.status, body);
    }

    throw new ApiRequestError(res.status, body);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json();
}

export function get<T>(path: string): Promise<T> {
  return request<T>(path, { method: "GET" });
}

export function post<T>(path: string, body?: unknown): Promise<T> {
  return request<T>(path, {
    method: "POST",
    body: body ? JSON.stringify(body) : undefined,
  });
}

export function put<T>(path: string, body?: unknown): Promise<T> {
  return request<T>(path, {
    method: "PUT",
    body: body ? JSON.stringify(body) : undefined,
  });
}

// Account Balances API functions (002-account-balances)
import {
  DepositRequest,
  WithdrawRequest,
  TransactionResponse,
  BalanceResponse,
  TransactionListResponse
} from "./types";

export function deposit(childId: number, data: DepositRequest): Promise<TransactionResponse> {
  return post<TransactionResponse>(`/children/${childId}/deposit`, data);
}

export function withdraw(childId: number, data: WithdrawRequest): Promise<TransactionResponse> {
  return post<TransactionResponse>(`/children/${childId}/withdraw`, data);
}

export function getBalance(childId: number): Promise<BalanceResponse> {
  return get<BalanceResponse>(`/children/${childId}/balance`);
}

export function getTransactions(childId: number): Promise<TransactionListResponse> {
  return get<TransactionListResponse>(`/children/${childId}/transactions`);
}

// Interest API functions
import { SetInterestResponse } from "./types";

// Combined interest (rate + schedule) API function
export interface SetInterestRequest {
  interest_rate_bps: number;
  frequency?: Frequency;
  day_of_week?: number;
  day_of_month?: number;
}

export function setInterest(childId: number, data: SetInterestRequest): Promise<SetInterestResponse> {
  return put<SetInterestResponse>(`/children/${childId}/interest`, data);
}

import {
  AllowanceSchedule,
  Frequency,
  UpcomingAllowancesResponse,
} from "./types";

export function getUpcomingAllowances(childId: number): Promise<UpcomingAllowancesResponse> {
  return get<UpcomingAllowancesResponse>(`/children/${childId}/upcoming-allowances`);
}

// Child-scoped allowance API functions (006-account-management-enhancements)

export function getChildAllowance(childId: number): Promise<AllowanceSchedule | null> {
  return get<AllowanceSchedule | null>(`/children/${childId}/allowance`);
}

export interface SetChildAllowanceRequest {
  amount_cents: number;
  frequency: Frequency;
  day_of_week?: number;
  day_of_month?: number;
  note?: string;
}

export function setChildAllowance(childId: number, data: SetChildAllowanceRequest): Promise<AllowanceSchedule> {
  return put<AllowanceSchedule>(`/children/${childId}/allowance`, data);
}

export function deleteChildAllowance(childId: number): Promise<void> {
  return request<void>(`/children/${childId}/allowance`, { method: "DELETE" });
}

export function pauseChildAllowance(childId: number): Promise<AllowanceSchedule> {
  return post<AllowanceSchedule>(`/children/${childId}/allowance/pause`);
}

export function resumeChildAllowance(childId: number): Promise<AllowanceSchedule> {
  return post<AllowanceSchedule>(`/children/${childId}/allowance/resume`);
}

// Interest schedule API functions (006-account-management-enhancements)
import { InterestSchedule } from "./types";

export function getInterestSchedule(childId: number): Promise<InterestSchedule | null> {
  return get<InterestSchedule | null>(`/children/${childId}/interest-schedule`);
}

export function deleteChild(childId: number): Promise<void> {
  return request<void>(`/children/${childId}`, { method: "DELETE" });
}

// Child Settings API functions (017-child-visual-themes)

export interface UpdateChildThemeResponse {
  message: string;
  theme: string;
}

export function updateChildTheme(theme: string): Promise<UpdateChildThemeResponse> {
  return put<UpdateChildThemeResponse>("/child/settings/theme", { theme });
}

export interface UpdateChildAvatarResponse {
  message: string;
  avatar: string | null;
}

export function updateChildAvatar(avatar: string | null): Promise<UpdateChildAvatarResponse> {
  return put<UpdateChildAvatarResponse>("/child/settings/avatar", { avatar });
}

// Subscription API functions (024-stripe-subscription)
import { SubscriptionResponse } from "./types";

export function getSubscription(): Promise<SubscriptionResponse> {
  return get<SubscriptionResponse>("/subscription");
}

export function createCheckoutSession(priceLookupKey: string): Promise<{ checkout_url: string }> {
  return post<{ checkout_url: string }>("/subscription/checkout", { price_lookup_key: priceLookupKey });
}

export function createPortalSession(): Promise<{ portal_url: string }> {
  return post<{ portal_url: string }>("/subscription/portal");
}

// Savings Goals API functions (025-savings-goals)
import { SavingsGoalsResponse, SavingsGoal, AllocateResponse, DeleteGoalResponse } from "./types";

export function getSavingsGoals(childId: number): Promise<SavingsGoalsResponse> {
  return get<SavingsGoalsResponse>(`/children/${childId}/savings-goals`);
}

export function createSavingsGoal(childId: number, data: { name: string; target_cents: number; emoji?: string }): Promise<SavingsGoal> {
  return post<SavingsGoal>(`/children/${childId}/savings-goals`, data);
}

export function updateSavingsGoal(childId: number, goalId: number, data: { name?: string; target_cents?: number; emoji?: string }): Promise<SavingsGoal> {
  return put<SavingsGoal>(`/children/${childId}/savings-goals/${goalId}`, data);
}

export function deleteSavingsGoal(childId: number, goalId: number): Promise<DeleteGoalResponse> {
  return request<DeleteGoalResponse>(`/children/${childId}/savings-goals/${goalId}`, { method: "DELETE" });
}

export function allocateToGoal(childId: number, goalId: number, amountCents: number): Promise<AllocateResponse> {
  return post<AllocateResponse>(`/children/${childId}/savings-goals/${goalId}/allocate`, { amount_cents: amountCents });
}

export function getGoalAllocations(childId: number, goalId: number): Promise<{ allocations: import("./types").GoalAllocation[] }> {
  return get<{ allocations: import("./types").GoalAllocation[] }>(`/children/${childId}/savings-goals/${goalId}/allocations`);
}

// Contact API function
export function submitContact(body: string) {
  return post<{ status: string }>("/contact", { body });
}

// Settings API functions (013-parent-settings)
import { SettingsResponse, UpdateTimezoneResponse, UpdateBankNameResponse, NotificationPreferences } from "./types";

export function getSettings(): Promise<SettingsResponse> {
  return get<SettingsResponse>("/settings");
}

export function updateTimezone(timezone: string): Promise<UpdateTimezoneResponse> {
  return put<UpdateTimezoneResponse>("/settings/timezone", { timezone });
}

export function updateBankName(bankName: string): Promise<UpdateBankNameResponse> {
  return put<UpdateBankNameResponse>("/settings/bank-name", { bank_name: bankName });
}

export function deleteAccount(): Promise<void> {
  return request<void>("/account", { method: "DELETE" });
}

// Notification Preferences API functions (033-email-notifications)

export function getNotificationPrefs(): Promise<NotificationPreferences> {
  return get<NotificationPreferences>("/settings/notifications");
}

export function updateNotificationPrefs(prefs: Partial<NotificationPreferences>): Promise<NotificationPreferences & { message: string }> {
  return put<NotificationPreferences & { message: string }>("/settings/notifications", prefs);
}

// Chore System API functions (031-chore-system)
import {
  Chore,
  ChoreInstance,
  ChoreListResponse,
  ChoreInstanceListResponse,
  PendingInstancesResponse,
  ApproveResponse,
  ChoreEarningsResponse,
  ChoreRecurrence,
} from "./types";

export interface CreateChoreRequest {
  name: string;
  description?: string;
  reward_cents: number;
  recurrence: ChoreRecurrence;
  day_of_week?: number;
  day_of_month?: number;
  child_ids: number[];
}

export function getChores(): Promise<ChoreListResponse> {
  return get<ChoreListResponse>("/chores");
}

export function createChore(data: CreateChoreRequest): Promise<{ chore: Chore }> {
  return post<{ chore: Chore }>("/chores", data);
}

export function getChildChores(): Promise<ChoreInstanceListResponse> {
  return get<ChoreInstanceListResponse>("/child/chores");
}

export function completeChore(instanceId: number): Promise<{ instance: ChoreInstance }> {
  return post<{ instance: ChoreInstance }>(`/child/chores/${instanceId}/complete`);
}

export function getPendingChores(): Promise<PendingInstancesResponse> {
  return get<PendingInstancesResponse>("/chores/pending");
}

export function getCompletedChores(limit: number, offset: number): Promise<{ instances: ChoreInstance[]; total: number }> {
  return get<{ instances: ChoreInstance[]; total: number }>(`/chores/completed?limit=${limit}&offset=${offset}`);
}

export function approveChore(instanceId: number): Promise<ApproveResponse> {
  return post<ApproveResponse>(`/chore-instances/${instanceId}/approve`);
}

export function rejectChore(instanceId: number, reason?: string): Promise<{ instance: ChoreInstance }> {
  return post<{ instance: ChoreInstance }>(`/chore-instances/${instanceId}/reject`, reason ? { reason } : {});
}

export function getChoreEarnings(): Promise<ChoreEarningsResponse> {
  return get<ChoreEarningsResponse>("/child/chores/earnings");
}

export function activateChore(choreId: number): Promise<{ chore: Chore }> {
  return request<{ chore: Chore }>(`/chores/${choreId}/activate`, { method: "PATCH" });
}

export function deactivateChore(choreId: number): Promise<{ chore: Chore }> {
  return request<{ chore: Chore }>(`/chores/${choreId}/deactivate`, { method: "PATCH" });
}

export function updateChore(choreId: number, data: Partial<CreateChoreRequest>): Promise<{ chore: Chore }> {
  return put<{ chore: Chore }>(`/chores/${choreId}`, data);
}

export function deleteChore(choreId: number): Promise<void> {
  return request<void>(`/chores/${choreId}`, { method: "DELETE" });
}

// Withdrawal Requests API functions (032-withdrawal-requests)
import {
  WithdrawalRequestSubmitRequest,
  WithdrawalRequestResponse,
  WithdrawalRequestListResponse,
  WithdrawalRequestApproveRequest,
  WithdrawalRequestApproveResponse,
  WithdrawalRequestDenyRequest,
  WithdrawalRequestPendingCountResponse,
} from "./types";

export function submitWithdrawalRequest(data: WithdrawalRequestSubmitRequest): Promise<WithdrawalRequestResponse> {
  return post<WithdrawalRequestResponse>("/child/withdrawal-requests", data);
}

export function getChildWithdrawalRequests(status?: string): Promise<WithdrawalRequestListResponse> {
  const query = status ? `?status=${status}` : "";
  return get<WithdrawalRequestListResponse>(`/child/withdrawal-requests${query}`);
}

export function cancelWithdrawalRequest(requestId: number): Promise<WithdrawalRequestResponse> {
  return post<WithdrawalRequestResponse>(`/child/withdrawal-requests/${requestId}/cancel`);
}

export function getWithdrawalRequests(params?: { status?: string; child_id?: number }): Promise<WithdrawalRequestListResponse> {
  const searchParams = new URLSearchParams();
  if (params?.status) searchParams.set("status", params.status);
  if (params?.child_id) searchParams.set("child_id", String(params.child_id));
  const query = searchParams.toString() ? `?${searchParams.toString()}` : "";
  return get<WithdrawalRequestListResponse>(`/withdrawal-requests${query}`);
}

export function approveWithdrawalRequest(requestId: number, data?: WithdrawalRequestApproveRequest): Promise<WithdrawalRequestApproveResponse> {
  return post<WithdrawalRequestApproveResponse>(`/withdrawal-requests/${requestId}/approve`, data || {});
}

export function denyWithdrawalRequest(requestId: number, data?: WithdrawalRequestDenyRequest): Promise<WithdrawalRequestResponse> {
  return post<WithdrawalRequestResponse>(`/withdrawal-requests/${requestId}/deny`, data || {});
}

export function getPendingWithdrawalRequestCount(): Promise<WithdrawalRequestPendingCountResponse> {
  return get<WithdrawalRequestPendingCountResponse>("/withdrawal-requests/pending/count");
}

