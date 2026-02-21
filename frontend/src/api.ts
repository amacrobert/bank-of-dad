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

// Settings API functions (013-parent-settings)
import { SettingsResponse, UpdateTimezoneResponse } from "./types";

export function getSettings(): Promise<SettingsResponse> {
  return get<SettingsResponse>("/settings");
}

export function updateTimezone(timezone: string): Promise<UpdateTimezoneResponse> {
  return put<UpdateTimezoneResponse>("/settings/timezone", { timezone });
}

export function deleteAccount(): Promise<void> {
  return request<void>("/account", { method: "DELETE" });
}

