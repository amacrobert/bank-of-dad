# Feature Specification: Child Visual Themes

**Feature Branch**: `017-child-visual-themes`
**Created**: 2026-02-18
**Status**: Draft
**Input**: User description: "As a child, I want to be able to change my visual theme. Add a child settings page, similar to the parent settings page. In that settings page, add an option to select a theme. The current (and default) theme will be called 'Sapling'. There will be two other themes for now: 'Piggy Bank' and 'Rainbow'. Different themes change only colors and background images. Each theme has 2 colors for copy text and background. The Sapling theme's copy text color will remain --color-forest, and its background color will remain bg-cream. Add a background image that's an svg. It should not be loud and add to the visual feel without commanding attention. Come up with colors and svg images for the other two themes too."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Select a Visual Theme (Priority: P1)

As a child user, I want to choose a visual theme from a settings page so that the app feels personalized and fun for me.

The child navigates to a new "Settings" page accessible from the sidebar/navigation. On that page, they see the three available themes displayed as visual previews. They tap one to select it. The entire child experience immediately updates to reflect the chosen theme's colors and background image. The selected theme is remembered so it persists across sessions.

**Why this priority**: This is the core feature — without theme selection, nothing else matters. It delivers the primary value of personalization.

**Independent Test**: Can be fully tested by logging in as a child, navigating to settings, selecting each theme, and verifying the visual appearance changes accordingly.

**Acceptance Scenarios**:

1. **Given** a child is logged in and on the settings page, **When** they view the theme options, **Then** they see three themes displayed: "Sapling", "Piggy Bank", and "Rainbow", each with a visual preview of its colors.
2. **Given** a child is on the settings page, **When** they select the "Piggy Bank" theme, **Then** the app's copy text color, background color, and background image all update to match the Piggy Bank theme immediately.
3. **Given** a child has previously selected "Rainbow" as their theme, **When** they log out and log back in, **Then** the Rainbow theme is still applied.
4. **Given** a child has never changed their theme, **When** they first visit the app, **Then** the "Sapling" theme is applied by default.

---

### User Story 2 - Child Settings Page Navigation (Priority: P2)

As a child user, I want a "Settings" link in my navigation so I can easily find and access theme preferences.

A new "Settings" item appears in the child's sidebar (desktop) and bottom tab bar (mobile), following the same navigation patterns used by the parent settings page. The settings page uses the same two-column layout as the parent settings page, with a category sidebar on the left and content on the right.

**Why this priority**: The settings page is the container for the theme feature. Without accessible navigation, children cannot discover or use theme selection.

**Independent Test**: Can be tested by logging in as a child and verifying the "Settings" navigation item appears and routes to the child settings page.

**Acceptance Scenarios**:

1. **Given** a child is logged in, **When** they view the navigation sidebar (desktop) or bottom tab bar (mobile), **Then** they see a "Settings" option alongside existing items like "Home" and "Growth".
2. **Given** a child taps the "Settings" navigation item, **When** the page loads, **Then** they see a settings page with a category navigation and content area, matching the structure of the parent settings page.

---

### User Story 3 - Theme Visual Design (Priority: P1)

As a child user, I want each theme to have a distinct, appealing visual identity so that switching themes feels meaningful and delightful.

Each theme defines two colors (copy text and background) plus a subtle SVG background image. The background image should enhance the visual atmosphere without being distracting — it should feel like gentle wallpaper, not a busy pattern.

**Why this priority**: Tied with P1 because the quality and distinctiveness of each theme is what makes the feature compelling. Generic or unappealing themes would undermine the entire feature.

**Independent Test**: Can be tested by applying each theme and verifying the correct colors and background image are rendered.

**Acceptance Scenarios**:

1. **Given** the "Sapling" theme is active, **Then** the copy text color is forest green, the background is cream, and a subtle nature-inspired SVG pattern (such as small scattered leaves or gentle vine tendrils) is visible in the background.
2. **Given** the "Piggy Bank" theme is active, **Then** the copy text color is a warm rose/mauve, the background is a soft blush pink, and a subtle money-inspired SVG pattern (such as faint coins or small piggy bank silhouettes) is visible in the background.
3. **Given** the "Rainbow" theme is active, **Then** the copy text color is a rich indigo/purple, the background is a soft lavender, and a subtle whimsical SVG pattern (such as tiny scattered stars or gentle rainbow arcs) is visible in the background.

---

### Edge Cases

- What happens if a child's saved theme is removed in a future update? The system falls back to the default "Sapling" theme.
- What happens if the theme preference fails to save? The child sees an error message and the previously saved theme remains active.
- What if a parent logs in — do they see the child's theme? No. Themes apply only to the child experience. Parents always see the standard app appearance.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a child settings page accessible from the child navigation.
- **FR-002**: The system MUST display three theme options on the child settings page: "Sapling" (default), "Piggy Bank", and "Rainbow".
- **FR-003**: Each theme option MUST show a visual preview so the child can see what the theme looks like before selecting it.
- **FR-004**: The system MUST apply the selected theme's colors and background image to all child-facing pages immediately upon selection.
- **FR-005**: The system MUST persist the child's theme preference so it survives logout and re-login.
- **FR-006**: The system MUST default to the "Sapling" theme for children who have not yet chosen a theme.
- **FR-007**: The system MUST NOT apply child themes to any parent-facing pages.
- **FR-008**: Each theme MUST define exactly two colors (copy text and background) and one subtle SVG background image.
- **FR-009**: The "Sapling" theme MUST use the app's existing forest green copy text color and warm cream background color, matching the current default appearance.
- **FR-010**: The SVG background images MUST be subtle and non-distracting — they should add to the visual feel without commanding attention.
- **FR-011**: The child settings page MUST follow the same layout structure as the parent settings page (category navigation + content area).

### Theme Definitions

| Theme      | Copy Text Color        | Background Color      | Background Image Description                              |
|------------|------------------------|-----------------------|-----------------------------------------------------------|
| Sapling    | Forest green (current) | Cream (current)       | Subtle scattered leaves or gentle vine tendrils           |
| Piggy Bank | Warm rose/mauve        | Soft blush pink       | Faint coins or small piggy bank silhouettes               |
| Rainbow    | Rich indigo/purple     | Soft lavender         | Tiny scattered stars or gentle rainbow arcs               |

### Key Entities

- **Theme**: A named visual configuration consisting of a display name, copy text color, background color, and SVG background image identifier.
- **Child Theme Preference**: A per-child record storing which theme the child has selected, defaulting to "Sapling" if none is set.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of child users see the "Sapling" theme by default on first login without any configuration.
- **SC-002**: Children can select and apply a new theme in under 10 seconds from opening the settings page.
- **SC-003**: Theme changes are visually apparent and apply to all child pages within 1 second of selection.
- **SC-004**: Theme preference persists across 100% of login sessions — a child who selected "Rainbow" sees "Rainbow" every time they return.
- **SC-005**: 0% of parent-facing pages are affected by any child theme selection.
- **SC-006**: All three themes are visually distinct from each other, with clearly different color palettes and background imagery.

## Assumptions

- The three themes ("Sapling", "Piggy Bank", "Rainbow") are the only themes needed for launch. More themes may be added later but are out of scope.
- Theme preference is stored per child, not per device or browser session.
- "All child-facing pages" includes the child dashboard, growth page, and the child settings page itself.
- The SVG background images are decorative patterns, not photographs or complex illustrations.
- Theme selection does not require parent approval.
- The child settings page will initially contain only the theme selection category. Additional settings categories may be added in the future.

## Scope Boundaries

### In Scope

- Child settings page with navigation integration
- Three theme definitions (colors + SVG backgrounds)
- Theme selection UI with visual previews
- Persisting theme preference per child
- Applying theme across all child-facing pages

### Out of Scope

- Custom/user-created themes
- Theme scheduling (e.g., dark mode at night)
- Animated themes or transitions between themes beyond the initial application
- Parent ability to restrict or control child theme choices
- Theme effects on fonts, icons, or layout — only colors and background images change
