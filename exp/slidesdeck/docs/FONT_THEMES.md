# Font Themes

Slidesdeck includes **16 professionally curated font themes** that pair Google Fonts for headings, body text, and code blocks. Each theme is designed to give your presentations a distinct personality while maintaining excellent readability.

## Quick Usage

### CLI Flag

Set the default font theme when generating your presentation:

```bash
slidesdeck --fonttheme elegant presentation.md
slidesdeck -f startup slides.org
```

### Web UI

Press `T` (Shift+t) during your presentation to open the Font Theme palette. Navigate with arrow keys, search by name or category, and press Enter to apply.

## Available Font Themes

### Modern (default)
- **Style**: Clean sans-serif
- **Fonts**: Inter (headings, body), JetBrains Mono (code)
- **Best for**: Technical presentations, startups, general purpose
- **Category**: Sans-serif

### Elegant
- **Style**: Sophisticated serif headings with clean body
- **Fonts**: Playfair Display (headings), Source Sans 3 (body), Fira Code (code)
- **Best for**: Formal presentations, fashion, luxury brands
- **Category**: Mixed

### Classic
- **Style**: Timeless serif throughout
- **Fonts**: Libre Baskerville (headings, body), Source Code Pro (code)
- **Best for**: Academic papers, literature, traditional industries
- **Category**: Serif

### Tech
- **Style**: Monospace-inspired technical aesthetic
- **Fonts**: Space Grotesk (headings, body), JetBrains Mono (code)
- **Best for**: Developer talks, API documentation, technical deep-dives
- **Category**: Mono

### Minimal
- **Style**: Ultra-clean geometric sans-serif
- **Fonts**: DM Sans (headings, body), DM Mono (code)
- **Best for**: Design presentations, portfolios, minimalist content
- **Category**: Sans-serif

### Editorial
- **Style**: Magazine-style high contrast
- **Fonts**: Oswald (headings), Merriweather (body), Roboto Mono (code)
- **Best for**: News-style presentations, bold statements, media
- **Category**: Mixed

### Friendly
- **Style**: Warm and approachable rounded fonts
- **Fonts**: Nunito (headings, body), IBM Plex Mono (code)
- **Best for**: Educational content, workshops, team presentations
- **Category**: Sans-serif

### Corporate
- **Style**: Professional business presentation style
- **Fonts**: Work Sans (headings, body), Ubuntu Mono (code)
- **Best for**: Business reports, quarterly reviews, corporate training
- **Category**: Sans-serif

### Academic
- **Style**: Scholarly serif fonts
- **Fonts**: Crimson Text (headings, body), Inconsolata (code)
- **Best for**: Research presentations, thesis defenses, scientific papers
- **Category**: Serif

### Creative
- **Style**: Artistic and expressive typography
- **Fonts**: Bebas Neue (headings), Lato (body), JetBrains Mono (code)
- **Best for**: Creative portfolios, art direction, bold storytelling
- **Category**: Mixed

### Luxury
- **Style**: High-end elegant typography
- **Fonts**: Cormorant Garamond (headings), Montserrat (body), Fira Code (code)
- **Best for**: Premium brands, high fashion, upscale services
- **Category**: Mixed

### Startup
- **Style**: Contemporary tech startup aesthetic
- **Fonts**: Manrope (headings, body), JetBrains Mono (code)
- **Best for**: Pitch decks, product launches, growth presentations
- **Category**: Sans-serif

### Vintage
- **Style**: Retro-inspired classic typography
- **Fonts**: Abril Fatface (headings), Lora (body), Courier Prime (code)
- **Best for**: Historical topics, retro branding, classic aesthetics
- **Category**: Mixed

### Swiss
- **Style**: International style clean typography
- **Fonts**: Helvetica Neue / Helvetica / Arial (system fonts)
- **Best for**: International business, clean documentation, timeless design
- **Category**: Sans-serif
- **Note**: Uses system fonts, no external loading required

### Playful
- **Style**: Fun and energetic for casual talks
- **Fonts**: Fredoka (headings), Quicksand (body), Nova Mono (code)
- **Best for**: Casual presentations, games, children's content, creative talks
- **Category**: Mixed

### Brutalist
- **Style**: Bold raw typography with strong contrast
- **Fonts**: Archivo Black (headings), Space Grotesk (body), Share Tech Mono (code)
- **Best for**: Bold statements, unconventional designs, attention-grabbing content
- **Category**: Mixed

## Category Reference

| Category | Description | Themes |
|----------|-------------|--------|
| **sans** | Clean, modern sans-serif fonts | modern, minimal, friendly, corporate, startup, swiss |
| **serif** | Traditional, readable serif fonts | classic, academic |
| **mixed** | Combination of serif and sans-serif | elegant, editorial, creative, luxury, vintage, playful, brutalist |
| **mono** | Monospace or technical aesthetic | tech |

## Technical Details

### Google Fonts Loading

Font themes load Google Fonts dynamically when selected. The first time you choose a theme, its fonts are fetched from Google Fonts CDN. Once loaded, they're cached in the browser for the session.

### CSS Variables

Font themes use CSS custom properties that you can override in custom CSS:

```css
:root {
  --font-heading: 'Your Heading Font', sans-serif;
  --font-body: 'Your Body Font', sans-serif;
  --font-code: 'Your Mono Font', monospace;
}
```

### Fallbacks

All font themes include system font fallbacks to ensure text remains readable if external fonts fail to load:

- **Sans-serif**: `system-ui, -apple-system, BlinkMacSystemFont, sans-serif`
- **Serif**: `Georgia, 'Times New Roman', serif`
- **Monospace**: `'SF Mono', Monaco, 'Courier New', monospace`

## Combining with Color Themes

Font themes work independently from daisyUI color themes. You can mix and match:

```bash
# Dark luxury theme
slidesdeck --theme luxury --fonttheme elegant presentation.md

# Light tech theme
slidesdeck --theme light --fonttheme tech presentation.md

# Corporate dark mode
slidesdeck --theme business --fonttheme corporate presentation.md
```

## Keyboard Shortcuts

| Key            | Action                             |
|----------------|------------------------------------|
| `T`            | Open Font Theme palette            |
| `↑` / `↓`      | Navigate themes                    |
| `Enter`        | Select highlighted theme           |
| `Esc`          | Close palette (revert to original) |
| Type to search | Filter themes by name or category  |
