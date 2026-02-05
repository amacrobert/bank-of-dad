import { ApiError } from "./types";

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
  const res = await fetch(`/api${path}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    ...init,
  });

  if (!res.ok) {
    let body: ApiError;
    try {
      body = await res.json();
    } catch {
      body = { error: "Unknown error" };
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

// Allowance Scheduling API functions (003-allowance-scheduling)
import {
  AllowanceSchedule,
  CreateScheduleRequest,
  UpdateScheduleRequest,
  ScheduleListResponse,
  UpcomingAllowancesResponse,
} from "./types";

export function createSchedule(data: CreateScheduleRequest): Promise<AllowanceSchedule> {
  return post<AllowanceSchedule>("/schedules", data);
}

export function listSchedules(): Promise<ScheduleListResponse> {
  return get<ScheduleListResponse>("/schedules");
}

export function getSchedule(id: number): Promise<AllowanceSchedule> {
  return get<AllowanceSchedule>(`/schedules/${id}`);
}

export function updateSchedule(id: number, data: UpdateScheduleRequest): Promise<AllowanceSchedule> {
  return put<AllowanceSchedule>(`/schedules/${id}`, data);
}

export function deleteSchedule(id: number): Promise<void> {
  return request<void>(`/schedules/${id}`, { method: "DELETE" });
}

export function pauseSchedule(id: number): Promise<AllowanceSchedule> {
  return post<AllowanceSchedule>(`/schedules/${id}/pause`);
}

export function resumeSchedule(id: number): Promise<AllowanceSchedule> {
  return post<AllowanceSchedule>(`/schedules/${id}/resume`);
}

export function getUpcomingAllowances(childId: number): Promise<UpcomingAllowancesResponse> {
  return get<UpcomingAllowancesResponse>(`/children/${childId}/upcoming-allowances`);
}
