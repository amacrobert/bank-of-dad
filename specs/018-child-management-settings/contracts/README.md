# API Contracts: 018-child-management-settings

## No New Contracts

This feature requires **no new API endpoints** and **no changes to existing endpoints**. All child CRUD operations use existing backend APIs:

| Method | Path | Used By |
|--------|------|---------|
| GET | /api/children | `ChildList` (Settings + Dashboard) |
| POST | /api/children | `AddChildForm` (Settings + Onboarding) |
| PUT | /api/children/{id}/name | `ChildAccountSettings` (Settings) |
| PUT | /api/children/{id}/password | `ChildAccountSettings` (Settings) |
| DELETE | /api/children/{id} | `ChildAccountSettings` (Settings) |

The "Used By" column reflects the new locations after this feature is implemented.
