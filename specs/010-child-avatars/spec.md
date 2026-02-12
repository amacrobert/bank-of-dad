# Feature Specification: Child Avatars

**Feature Branch**: `010-child-avatars`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "Add avatars to children accounts. As a parent, I want to be able to select an avatar for each of my children to quickly differentiate them in the selection list and add a cute visual flourish. Use a limited selection of emojis as the avatars available to pick. If a child has an avatar set, it should be displayed in the Avatar circle instead of the letter of their first name. Parents should be able to change their child's avatar in the 'Update name' form (rename this form 'Update name and avatar')."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Set avatar when creating a child (Priority: P1)

A parent is adding a new child to the family bank. Below the name input field, they see a grid of emoji avatars to choose from (e.g., sunflower, butterfly, star, mushroom). They tap one to select it — the selected emoji gets a visible highlight. After submitting the form, the child appears in the child list with the chosen emoji in their avatar circle instead of the first letter of their name.

**Why this priority**: This is the primary entry point for avatar selection and the most natural moment to pick one — when setting up the child's identity for the first time.

**Independent Test**: Can be fully tested by creating a new child with an avatar selected and verifying the avatar emoji appears in the child list.

**Acceptance Scenarios**:

1. **Given** a parent is on the Add Child form, **When** they view the form, **Then** they see a grid of emoji avatars below the name field.
2. **Given** a parent has not selected an avatar, **When** they submit the form, **Then** the child is created successfully with no avatar (the avatar circle shows the first letter of the child's name as it does today).
3. **Given** a parent has selected an avatar, **When** they submit the form, **Then** the child is created with the selected avatar and it appears in the child list avatar circle.
4. **Given** a parent has selected an avatar, **When** they tap a different avatar, **Then** the selection moves to the newly tapped avatar.

---

### User Story 2 - Display avatar throughout the app (Priority: P1)

Wherever a child's avatar circle currently displays the first letter of their name, it should instead display their emoji avatar if one is set. This includes the child selection list on the parent dashboard and the child login screen. If no avatar is set, the existing first-letter behavior remains unchanged.

**Why this priority**: Avatars only provide value if they are consistently displayed wherever children are visually represented.

**Independent Test**: Can be tested by creating a child with an avatar and navigating to each screen where the avatar circle appears, verifying the emoji is shown.

**Acceptance Scenarios**:

1. **Given** a child has an avatar set, **When** the parent views the child selection list, **Then** the avatar circle displays the emoji instead of the first letter.
2. **Given** a child has an avatar set, **When** the child logs in and sees their avatar circle on the login screen, **Then** the emoji is displayed.
3. **Given** a child does not have an avatar set, **When** the parent views the child selection list, **Then** the avatar circle displays the first letter of the child's name as before.

---

### User Story 3 - Update avatar in account settings (Priority: P2)

A parent wants to change their child's avatar (or set one for an existing child who doesn't have one yet). They open the child's account settings and find the "Update Name" form has been renamed to "Update Name and Avatar." The form now includes the same emoji grid from the Add Child form, pre-selecting the child's current avatar if one exists. The parent can select a different emoji or clear the selection, then submit the form to save.

**Why this priority**: Important for ongoing management but secondary to the initial creation flow, since most avatars will be set at creation time.

**Independent Test**: Can be tested by opening an existing child's account settings, changing their avatar, and verifying the new avatar appears in the child list.

**Acceptance Scenarios**:

1. **Given** a parent is viewing a child's account settings, **When** they see the name update form, **Then** it is titled "Update Name and Avatar" and includes the emoji avatar grid.
2. **Given** a child already has an avatar, **When** the parent opens the form, **Then** the child's current avatar is pre-selected in the grid.
3. **Given** a parent selects a new avatar and submits the form, **When** the update succeeds, **Then** the child's avatar is changed and the child list reflects the new avatar.
4. **Given** a parent deselects the current avatar (taps it again to clear), **When** they submit the form, **Then** the child's avatar is removed and the avatar circle reverts to showing the first letter of their name.

---

### Edge Cases

- What happens if the avatar emoji set is updated in a future release and a child has an emoji that is no longer in the set? The stored emoji should still display correctly; it simply won't appear as a selectable option for new assignments.
- What happens if the parent only changes the avatar but not the name? The update should succeed — neither field should be required to change for a successful submission.
- What happens if the parent only changes the name but not the avatar? The existing avatar should be preserved.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a fixed set of emoji avatars for selection: 🌻 🌿 🍂 🌸 🌊 🌙 ⭐ 🦋 🐝 🍄 🐸 🦊 🐻 🐰 🐢 🎨
- **FR-002**: The Add Child form MUST display the emoji avatar grid below the name input, allowing the parent to optionally select one avatar.
- **FR-003**: The avatar selection MUST visually highlight the currently selected emoji (e.g., a colored border or background).
- **FR-004**: Avatar selection MUST be optional — a child can be created or updated without an avatar.
- **FR-005**: The "Update Name" form in Account Settings MUST be renamed to "Update Name and Avatar" and MUST include the emoji avatar grid.
- **FR-006**: The avatar grid in the update form MUST pre-select the child's current avatar if one exists.
- **FR-007**: A parent MUST be able to clear a child's avatar by tapping the currently selected emoji to deselect it.
- **FR-008**: Wherever the child's avatar circle displays (child selection list, login screen), the system MUST show the emoji avatar if set, or the first letter of the child's name if not set.
- **FR-009**: The avatar MUST be stored as a single emoji character associated with the child's record.
- **FR-010**: Updating only the avatar (without changing the name) MUST be allowed and MUST succeed.

### Key Entities

- **Child Account**: Existing entity. Gains a new optional attribute: avatar (a single emoji character, nullable). When present, the avatar is displayed in place of the name initial.
- **Avatar Set**: A fixed list of emoji characters available for selection. Defined in the application, not user-configurable.

## Assumptions

- The emoji avatar set is hardcoded in the frontend and does not need to be served from the backend. The backend stores and returns the emoji string but does not validate it against the allowed set (the frontend controls what's selectable).
- Avatars are purely cosmetic and have no impact on authentication, permissions, or financial features.
- The same emoji can be used by multiple children in the same family — there is no uniqueness constraint on avatars.
- The avatar grid layout should wrap naturally on smaller screens (responsive grid).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A parent can select an avatar for a child during account creation in under 5 seconds (single tap on the emoji grid).
- **SC-002**: Avatars are consistently displayed in every location where the child's avatar circle appears — no location shows the first-letter fallback when an avatar is set.
- **SC-003**: Parents can change or remove a child's avatar from account settings without affecting any other account data.
- **SC-004**: Children created without an avatar continue to display the first-letter initial with no visual regression.
