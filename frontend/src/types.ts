export interface ParentUser {
  user_type: "parent";
  user_id: number;
  family_id: number;
  display_name: string;
  email: string;
  family_slug: string;
}

export interface ChildUser {
  user_type: "child";
  user_id: number;
  family_id: number;
  first_name: string;
  family_slug: string;
}

export type AuthUser = ParentUser | ChildUser;

export interface Child {
  id: number;
  first_name: string;
  is_locked: boolean;
  balance_cents: number;
  created_at: string;
}

export interface Family {
  id: number;
  slug: string;
}

export interface FamilyCheck {
  slug: string;
  exists: boolean;
}

export interface SlugCheck {
  slug: string;
  available: boolean;
  valid: boolean;
  suggestions: string[];
}

export interface ApiError {
  error: string;
  message?: string;
  suggestions?: string[];
}

export interface ChildCreateResponse {
  id: number;
  first_name: string;
  family_slug: string;
  login_url: string;
}

export interface ChildListResponse {
  children: Child[];
}

// Account Balances Feature (002-account-balances)

export type TransactionType = 'deposit' | 'withdrawal';

export interface Transaction {
  id: number;
  child_id: number;
  parent_id: number;
  amount_cents: number;
  type: TransactionType;
  note?: string;
  created_at: string;
}

export interface BalanceResponse {
  child_id: number;
  balance_cents: number;
}

export interface ChildWithBalance {
  id: number;
  first_name: string;
  balance_cents: number;
  is_locked: boolean;
}

export interface TransactionListResponse {
  transactions: Transaction[];
}

export interface TransactionResponse {
  transaction: Transaction;
  new_balance_cents: number;
}

export interface DepositRequest {
  amount_cents: number;
  note?: string;
}

export interface WithdrawRequest {
  amount_cents: number;
  note?: string;
}
