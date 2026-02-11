# Frontend Design Prompt: Children's Banking App (SPA)

## Vision

Update the single-page application in frontend/, redesigning it while keeping all existing functionality intact.
This is a children's banking/allowance platform called **"Bank of Dad"** — a family finance app where parents set up accounts and children learn to save, spend, and track their money.
The design should feel like a **premium fintech product that happens to be for families**, not a toy. Think **Monzo meets Headspace meets a beautifully illustrated children's book** — warm, trustworthy, and delightful without being cartoonish or condescending.

---

## Aesthetic Direction

**Tone: "Soft Modernist / Friendly Fintech"**

- **NOT** this: Garish primary colors, chunky cartoon fonts, mascot characters, rainbow gradients, clip-art style illustrations, "kid zone" energy, or anything that screams "educational software from 2012."
- **YES** this: Rounded geometric shapes, a nature-inspired palette (soft greens, warm ambers, creamy whites, gentle earth tones), generous whitespace, friendly but sophisticated typography, subtle depth through layered cards and soft shadows. Organic forms that evoke growth (leaves, rings of a tree, gentle curves) used as abstract decorative elements — never literal clip-art.

**Color Palette:**
- Use a **warm, earthy-yet-modern** palette. Think muted sage greens, soft terracotta, warm sand/cream backgrounds, with a confident teal or deep forest green as the primary action color. Accent with a cheerful golden amber for highlights, badges, and progress indicators. Avoid pure white — opt for warm off-whites and light creams.

**Typography:**
- Choose a **rounded geometric sans-serif** for headings that feels approachable but not babyish (think Nunito, Quicksand, Outfit, or Poppins — but pick ONE distinctive choice and commit). Pair with a clean, highly legible body font. Text should be generously sized — remember, 6-year-olds may be reading some of this. Minimum touch targets of 48px. Labels should be simple, clear words.

**Iconography & Illustration:**
- Use simple, **line-based or filled geometric icons** (not detailed illustrations). Think Phosphor Icons or Lucide style. For decorative elements, use abstract organic shapes — soft blobs, leaf silhouettes, gentle circular patterns — as background textures, not as foreground content.

**Motion & Micro-interactions:**
- Smooth, calm transitions between pages (no jarring cuts). Gentle scale-up on tap for interactive elements. Page loads should have staggered fade-in reveals. Nothing should flash, shake, or demand attention aggressively. The motion language should feel **calm and encouraging**, like a deep breath.

---

## Layout & Responsiveness

- **Mobile-first.** The primary experience is a phone in a child's or parent's hand. Design for 375px-width screens first, then scale gracefully up.
- On desktop, the app should center in a comfortable max-width container (~480px for child views, ~960px for parent dashboard) with the warm cream/sand background extending to the edges. The parent dashboard can use a wider two-column layout on desktop.
- Navigation should be a **bottom tab bar on mobile** (3-4 tabs max with large, labeled icons) and can shift to a **sidebar or top nav on desktop**.
- Every interactive element must be large, obvious, and easy to tap. Generous padding. No tiny text links. No ambiguous icons without labels.

---

## Pages to Build

### 1. **HomePage**
The landing/marketing page for unauthenticated users. This is the front door.
- Hero section with the app name "Bank of Dad", a warm tagline like *"Watch your family's savings grow"*, and a clear CTA to get started.
- Brief value props (3 max): Parents set the rules, kids learn by doing, everything in one place.
- A "Sign in with Google" button and a "Family Login" link for returning families.
- Subtle animated background element (floating leaf shapes, gently growing rings).
- Footer with minimal links.

### 2. **SetupPage**
Post-authentication onboarding for new families. A multi-step wizard.
- Step 1: Name your family (e.g., "The Johnsons")
- Step 2: Add children (name + avatar selection from a set of 8-10 friendly abstract avatars — geometric animals or plant-based icons, not photos)
- Step 3: Set initial allowance rules (amount, frequency) -- this will apply to all children. If there are multiple children, show a tip box informing that you can change allowance per child after setup.
- Step 4: Confirmation / "You're all set!" celebration screen with a subtle confetti or sparkle animation
- Progress indicator at top (dots or a segmented bar). Back/Next buttons always visible and large.

### 3. **GoogleCallback**
Minimal — just a centered loading spinner with the Sapling logo and "Signing you in..." text. Branded and calm, not a blank white screen. Auto-redirects.

### 4. **FamilyLogin**
For returning families where multiple children share a device.
- Shows the family name at top.
- Grid of **large, tappable avatar cards** — one per family member (children + parent). Each card shows the avatar icon and first name. Think of the Netflix profile selector but warmer and more tactile.
- Tapping a child's avatar goes to ChildDashboard. Tapping the parent avatar may prompt for a PIN or go to ParentDashboard.
- This page must be usable by a 5-year-old picking their own profile.

### 5. **ChildDashboard**
The core experience for kids. This should be the most polished, most delightful page.
- **Balance display** front and center — large, friendly number with a subtle animated coin/leaf icon. Cents are smaller and raised.
- **Recent activity** feed: Simple list of recent transactions with icons (allowance received ↓, spent ↑, bonus ⭐). Minimal detail, easy to scan.
- Bottom tab bar: Me (name), logout.
- The overall feel should make a kid feel **proud and empowered** looking at their money, not overwhelmed.

### 6. **ParentDashboard**
The control center for parents. More information-dense but still clean.
- **Family overview**: Cards for each child showing balance, recent activity summary, and pending requests.
- **Quick actions**: "Send Allowance", "Approve Request", "Add Bonus", "Edit Rules".
- **Allowance schedule** summary: Shows automated allowance settings with easy edit access.
- **Activity log**: Filterable by child, date, type.
- On desktop, use a two-column layout (child selector sidebar + detail area). On mobile, stack with a child-tab selector at top.
- This page should feel like a **calm, organized command center** — trustworthy and efficient, not cluttered.

### 7. **NotFound**
A friendly, on-brand 404 page.
- A whimsical illustration or animation (a little lost leaf blowing in the wind, or a sapling looking around confused — keep it abstract/geometric, not cartoony).
- Text: "Oops! This page wandered off." with a clear "Go Home" button.
- Should make people smile, not feel frustrated.

---

## Technical Notes

- Keep **React SPA** structure, continuing to use React Router for page navigation.
- Use **Tailwind CSS** utility classes for styling — but only core/pre-defined classes (no custom Tailwind config or JIT).
- Import fonts from Google Fonts via `<link>` or `@import`.
- Use **Lucide React** for icons.
- All data should some from the backend as it currently does.
- Ensure accessible contrast ratios, semantic HTML, and ARIA labels on interactive elements.
- Animate with CSS transitions/animations — keep it performant.

---

## Summary of What Makes This Great

1. **Respects the audience**: Kids get clarity, large targets, and delight. Parents get control and calm. Nobody gets patronized.
2. **Cohesive identity**: The "Sapling" nature/growth metaphor runs through color, shape, and motion without being heavy-handed.
3. **Mobile-first but desktop-beautiful**: Every pixel is considered at 375px, then gracefully enhanced.
4. **Premium feel**: This should look like it belongs next to Monzo, Revolut, or Mercury — not next to a Flash game from 2005.

---

## Testing strategy

Navigate to http://localhost:8000 and ensure that the frontend does not have any errors.
Use the docker compose CLI to rebuild the frontend container if necessary.
