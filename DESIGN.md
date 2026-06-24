# DESIGN.md

Design system documentation for TokenJoy — adapted from "Corporate Trust" for an enterprise admin dashboard context.

## Philosophy

TokenJoy's visual language combines **professional credibility** with **modern vibrancy**. As an LLM API management platform, the UI must handle dense data (tables, charts, metrics) while feeling warm and premium — not clinical.

Key adaptations for a dashboard context:

- 3D transforms and isometric effects are used sparingly (sidebar logo, hero stats only)
- Colored shadows remain but at lower intensity for tables/cards that appear in quantity
- Gradient text is reserved for page titles and key metrics — never on body text
- Atmospheric blur orbs appear only in the sidebar and empty states, not on data-heavy pages

## Design Tokens

### Colors (Tailwind CSS v4 custom properties)

| Token                  | Value                  | Usage                      |
| ---------------------- | ---------------------- | -------------------------- |
| `--background`         | `#F8FAFC` (Slate 50)   | Page background            |
| `--foreground`         | `#0F172A` (Slate 900)  | Primary text               |
| `--card`               | `#FFFFFF`              | Card/surface background    |
| `--primary`            | `#4F46E5` (Indigo 600) | Brand color, active states |
| `--primary-foreground` | `#FFFFFF`              | Text on primary            |
| `--secondary`          | `#7C3AED` (Violet 600) | Gradient endpoint, accents |
| `--muted`              | `#F1F5F9` (Slate 100)  | Subtle backgrounds         |
| `--muted-foreground`   | `#64748B` (Slate 500)  | Secondary text             |
| `--border`             | `#E2E8F0` (Slate 200)  | Borders, dividers          |
| `--accent`             | `#EEF2FF` (Indigo 50)  | Highlight backgrounds      |
| `--destructive`        | `#EF4444` (Red 500)    | Error/danger states        |

### Typography

- **Font**: `Plus Jakarta Sans` (variable weight)
- **Headings**: weight 700-800, tracking `-0.02em`, line-height 1.1
- **Body**: weight 400-500, line-height 1.6
- **Labels/Nav**: weight 500-600, `text-sm`
- **Scale**: Page title `text-xl`/`text-2xl`, section title `text-sm font-semibold`

### Shadows (Colored)

```css
--shadow-card: 0 1px 3px rgba(79, 70, 229, 0.04), 0 4px 12px rgba(79, 70, 229, 0.06);
--shadow-card-hover: 0 4px 16px rgba(79, 70, 229, 0.1), 0 8px 24px rgba(79, 70, 229, 0.06);
--shadow-button: 0 2px 8px rgba(79, 70, 229, 0.25);
--shadow-sidebar: 4px 0 24px rgba(79, 70, 229, 0.06);
```

### Radius

- Cards: `rounded-xl` (12px)
- Inputs: `rounded-lg` (8px)
- Buttons: `rounded-lg` (8px) for inline, `rounded-full` for CTAs
- Badges: `rounded-full`

### Gradients

- **Primary**: `from-indigo-600 to-violet-600` — buttons, active nav, key metrics
- **Background accent**: `from-indigo-50 to-violet-50` — stat card backgrounds
- **Text gradient**: `bg-gradient-to-r from-indigo-600 to-violet-600 bg-clip-text text-transparent` — page titles
- **Sidebar**: `from-slate-900 via-indigo-950 to-slate-900` — dark dramatic sidebar

## Component Patterns

### Sidebar

- Dark background with indigo-tinted gradient
- Active item: pill shape with indigo-500 background and glow
- Icons from lucide-react, `h-4 w-4`
- Group labels: uppercase, `text-[10px]`, tracking-wider, indigo-300 tint
- Subtle atmospheric orb at bottom (blur-3xl, low opacity)

### Header

- Clean white, minimal — just breadcrumb context + user avatar
- Subtle bottom border, no heavy separators
- User avatar with indigo ring on hover

### Cards (Data)

- White bg, `rounded-xl`, subtle indigo-tinted shadow
- `hover:-translate-y-0.5` + enhanced shadow (for clickable cards only)
- Section title: `text-sm font-semibold text-slate-700`
- Stat cards: gradient accent background with large bold number

### Tables

- Clean borders, `rounded-xl` container
- Header: `bg-slate-50/80` with medium weight text
- Rows: subtle hover `bg-indigo-50/30`
- No heavy alternating colors — rely on spacing

### Badges/Status

- Rounded-full, small text
- Status colors: green (active), yellow (pending), red (error), slate (disabled)
- Provider badges: outline style with brand color hint

### Buttons

- **Primary**: gradient bg (indigo→violet), white text, colored shadow, lift on hover
- **Secondary**: white bg, slate border, hover slate-50
- **Ghost**: no bg, text color only, hover bg-slate-50
- **Destructive**: red-500 text (ghost style in tables), solid red for confirmations

### Progress Bars

- Track: `bg-slate-100`
- Fill: gradient `from-indigo-500 to-violet-500`
- Height: `h-2` with `rounded-full`

## Layout

- Sidebar width: `w-60` (240px)
- Main content padding: `p-8`
- Card grid gap: `gap-6`
- Page title margin-bottom: `mb-8`
- Section spacing: `space-y-6`
- Max content width: none (fills available space in admin context)

## Motion

- Card hover: `transition-all duration-200 ease-out`
- Button hover: `transition-all duration-150`
- Sidebar nav: `transition-colors duration-150`
- Page transitions: none (instant route changes for admin speed)

## Iconography

- Library: `lucide-react`
- Nav icons: `h-4 w-4`, stroke-width 2
- Feature icons: `h-5 w-5` in `bg-indigo-50 text-indigo-600 rounded-lg p-2` container
- Always paired with text in navigation
