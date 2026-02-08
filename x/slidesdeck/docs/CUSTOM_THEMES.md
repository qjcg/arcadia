# Custom Themes and Font Themes

Slidesdeck allows you to provide your own CSS files to define custom color themes and font themes. These are injected into the generated HTML and can be activated via CLI flags or the runtime theme palettes.

## Custom Color Themes

To define a custom color theme, use the `[data-theme="your-theme-name"]` selector in your CSS file.

### Example: `custom-theme.css`

```css
[data-theme="my-brand"] {
  --color-base-100: #ffffff;
  --color-base-content: #1f2937;
  --color-primary: #3b82f6;
  --color-primary-content: #ffffff;
  /* Add other daisyUI / Tailwind variables as needed */
}
```

### Usage

```bash
slidesdeck --theme-file custom-theme.css --theme my-brand presentation.md
```

## Custom Font Themes

To define a custom font theme, use the `html[data-font-theme="your-font-theme-name"]` selector.

### Example: `custom-fonts.css`

```css
html[data-font-theme="corporate-special"] {
  --font-heading: 'Helvetica Neue', Helvetica, sans-serif;
  --font-body: 'Georgia', serif;
  --font-code: 'Courier New', monospace;
}
```

### Usage

```bash
slidesdeck --fonttheme-file custom-fonts.css --fonttheme corporate-special presentation.md
```

## How It Works

1.  **Injection**: Your custom CSS files are read and embedded directly into a `<style>` tag in the `<head>` of the generated HTML file.
2.  **Activation**: When you use the `--theme` or `--fonttheme` flags, Slidesdeck sets the corresponding `data-theme` or `data-font-theme` attribute on the `<html>` element.
3.  **Scoped Selectors**: By using the attribute selectors (`[data-theme=...]` and `html[data-font-theme=...]`), your styles only apply when that specific theme is active.

## Overriding Existing Styles

If you want to override styles globally without creating a new theme, you can simply use standard selectors like `:root`, `body`, or `.slide`:

```css
:root {
  --radius-box: 0; /* Remove rounded corners everywhere */
}

.slide h1 {
  text-decoration: underline;
}
```
