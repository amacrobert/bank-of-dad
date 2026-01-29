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
