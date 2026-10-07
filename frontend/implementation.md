# bAI Landing Page Implementation Plan

## 1. Visual Aesthetic & Theme
The design features a dark, highly polished, "zen/mindful" aesthetic. We will adapt the placeholder text (which references sleep/meditation) to strictly highlight **bAI's** features while maintaining the calming, focused, and elegant vibe.

- **Color Palette**:
  - **Backgrounds**: Deep forest greens and dark teals (e.g., `#040f13`, `#0a1c1d`, `#0b0d11`).
  - **Accents**: Mint green/emerald (`#81e6d9`, `#a7f3d0`).
  - **Text**: White for primary headings, light greenish-gray (`#94a3b8` or `#8ba6a5`) for descriptions.
  - **Cards & UI**: Glassmorphic overlays, semi-transparent dark backgrounds (`bg-white/5` or `bg-[#0d2222]/80`) with subtle 1px glowing borders (`border-white/10`).
- **Typography**:
  - **Headings**: An elegant Serif font (e.g., *Playfair Display* or *Instrument Serif*) to give that premium, editorial look ("Your Coding Agent in the Terminal.").
  - **Body & UI**: Clean Sans-Serif (Geist, Inter) for standard text.
  - **Terminal/Code**: Monospace (Geist Mono).

## 2. Component Structure

The page is structured as a continuous scrolling landing page divided into horizontal bands.

### A. `Navbar`
- **Left**: bAI Logo (Custom icon `>` + text).
- **Center**: Nav links (Home, Features, Sessions, Themes, Docs).
- **Right**: Search icon, "Get Started &rarr;" pill button (outline style).

### B. `HeroSection`
- **Left Column**:
  - Eyebrow: `ONE BINARY. ANY MODEL.`
  - Serif Headline: `Your Coding Agent\nin the Terminal.`
  - Subhead: `Read. Edit. Undo. One binary for every model.`
  - CTAs: "Install bAI &rarr;" (solid mint) and "Watch Demo" (play button + text).
- **Center/Right Background**: A serene landscape image (we will use a placeholder or generic Unsplash landscape representing focus/calm).
- **Right Floating Card (Glassmorphism)**:
  - Title: "Which model today?"
  - Subtitle: "Seamlessly swap between top models without restarting."
  - Row of icons representing: GPT, Claude, DeepSeek, Groq, Local.
  - Decorative wave/chart at the bottom.

### C. `FeatureBar`
- A wide, bordered, pill-shaped container spanning the screen width.
- Contains 5 distinct feature highlights separated by subtle vertical dividers:
  1. **Shadow Git**: Invisible undo history.
  2. **Sessions**: SQLite-backed history.
  3. **Themes**: Hot-swappable colors.
  4. **Skills**: Global & local SKILL.md.
  5. **Notifications**: Herdr & audio alerts.

### D. `EverydayCoding` (Split Section)
- **Left**: Serif Headline "Built for everyday coding." + description + "Explore All Features &rarr;" CTA.
- **Center/Right UI Showcase**:
  - A large central glass panel showing a terminal/TUI mockup of bAI running.
  - A vertical stack of 3 smaller interaction cards on the right (Sessions, Themes, Shadow Git) showing active states.

### E. `TestimonialSection`
- Right-aligned text over a beautiful nature background.
- Eyebrow: `SMALL STEPS. BIG CHANGES.`
- Serif Headline: `A cleaner repo builds better code.`
- Quote text with user profile picture and left/right navigation arrows.

### F. `FooterShowcase` (Bottom Split)
- **Left**: "Undo Anything. Keep Shipping." + CTA "Explore Shadow Git".
- **Center**: A mini-carousel of 3 landscape cards representing features (e.g., Safe Experimentation, Atomic Edits, Easy Rollbacks).
- **Right**: "Built for developer productivity."
  - Stats Grid: 1 (Binary), 8+ (Providers), SQLite (DB), 50+ (Tools).

## 3. Tech Stack & Dependencies
- **Framework**: React 19 + Vite.
- **Styling**: Tailwind CSS v4.
- **Icons**: `lucide-react`.
- **Utilities**: `clsx`, `tailwind-merge` for class deduplication.
- **Fonts**: Add `Playfair Display` (Serif) via Google Fonts in `index.html`.

## 4. Execution Steps

1. **Step 1: Setup & Assets**
   - Update `index.html` to include the required Serif font.
   - Configure basic layout shell, background colors, and global CSS.
2. **Step 2: Navbar & Hero**
   - Build the top navigation.
   - Build the two-column Hero section with the floating glassmorphic card.
3. **Step 3: Feature Bar**
   - Build the 5-column horizontal feature strip.
4. **Step 4: Everyday Coding Section**
   - Build the asymmetric grid with the text on the left and the UI cards on the right.
5. **Step 5: Testimonial & Bottom Showcase**
   - Build the right-aligned testimonial.
   - Build the bottom grid with stats and the mini card carousel.
