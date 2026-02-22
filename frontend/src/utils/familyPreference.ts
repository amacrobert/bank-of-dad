const FAMILY_SLUG_KEY = "family_slug";

export function getFamilySlug(): string | null {
  return localStorage.getItem(FAMILY_SLUG_KEY);
}

export function setFamilySlug(slug: string): void {
  localStorage.setItem(FAMILY_SLUG_KEY, slug);
}

export function clearFamilySlug(): void {
  localStorage.removeItem(FAMILY_SLUG_KEY);
}
