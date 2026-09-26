# MathTrail — site (static export)

A static build of the MathTrail home page for parents, in English and Russian. No build step, no server code, no requests to other domains.

## Open it

- Locally: open `index.html` (it forwards to `en/`), or serve the folder: `python3 -m http.server 8080` and go to http://localhost:8080/.
- GitHub Pages: publish this folder as the site root.

## Layout

```
index.html                forwards to en/ (URL scheme of T19: /<locale>/, /<locale>/privacy/, /<locale>/terms/)
en/index.html             home page, English (authoritative)
ru/index.html             home page, Russian
en|ru/privacy/, terms/    placeholders until T19a
assets/tokens.css         design tokens (light, dark, and dark by prefers-color-scheme)
assets/mathtrail.css|.js  the widget components (window.MathTrail) from the design system
assets/site.css|.js       the page's layout and behaviour
assets/vendor/            React 18.3.1 and ReactDOM 18.3.1 (MIT)
```

## How it behaves

- All text is plain HTML and reads with JavaScript switched off; the six demo cards are pre-rendered and then brought to life (hydrated) by React.
- On a wide screen the card on the right stays in view and changes as the lesson steps scroll past; below 960px each step carries its own card.
- Light or dark follows the visitor's system setting.
- Language switch = links between `/en/` and `/ru/`.

## Placeholders to fill

- `[DOMAIN]` — in the connector address, `canonical`, `hreflang` and Open Graph URLs.
- `[CONTACT EMAIL]` — footer.
- Privacy policy and terms — T19a.
- Demo task: the traps for options A, D and E are illustrative; T66a asks for the demo to come from a reference task in `content/`.
- Open Graph image and `sitemap.xml`, `robots.txt`, `CNAME` — not included yet.
