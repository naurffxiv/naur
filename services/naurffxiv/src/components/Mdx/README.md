# MDX Components

Custom components available in all `.mdx` fight-guide pages.
They are registered in [`Layout/MdxComponents.js`](./Layout/MdxComponents.js) and require no import in the MDX file itself.

---

## Layout / Automatic Behaviours

### Images (`![alt](src)`)

Standard markdown images are automatically routed through `ImageModal`:

- Displayed as a clickable thumbnail; clicking opens the full-size image in a modal.
- The thumbnail src has its extension swapped to `.avif`. Make sure an `.avif` version of every image exists (run the sharp conversion script if needed).
- `width` / `height` are injected at build time by `rehype-img-size` — no need to specify them manually.

### External links (`[text](https://...)`)

Any `href` that starts with `http` automatically gets `target="_blank"`.

### Last Updated

Rendered automatically right after the first `# H1` heading on every page. It reads the file's git modification time — no action needed in the MDX.

---

## Elements

### `<Banner>`

A full-width hero image at the top of a page.

| Prop   | Type    | Default  | Description                                        |
| ------ | ------- | -------- | -------------------------------------------------- |
| `src`  | string  | required | Path to the image (relative to `/public`)          |
| `alt`  | string  | required | Alt text                                           |
| `left` | boolean | `false`  | Align the image to the left edge instead of center |

```mdx
<Banner
  src="/images/ultimates/kefka/kefka-banner.webp"
  alt="Dancing Mad (Ultimate)"
/>
```

---

### `<Callout>`

A highlighted callout box. Defaults to "Note" style if `type` is omitted.

| Prop   | Type   | Default     | Description                                                |
| ------ | ------ | ----------- | ---------------------------------------------------------- |
| `type` | string | `undefined` | `"warning"`, `"tip"`, or omit for the default "Note" style |

```mdx
<Callout>A general note with no type.</Callout>

<Callout type="warning">Watch out for something!</Callout>

<Callout type="tip">Here's a helpful tip.</Callout>
```

---

### `<Details>`

A collapsible dropdown. Optionally links to an external resource and/or provides a fullscreen button for an embedded iframe.

| Prop         | Type    | Default  | Description                                                                                                          |
| ------------ | ------- | -------- | -------------------------------------------------------------------------------------------------------------------- |
| `title`      | string  | required | Text shown in the summary row                                                                                        |
| `href`       | string  | —        | If set, shows an ↗ icon that opens the URL in a new tab                                                              |
| `fullscreen` | boolean | —        | If set, shows a ⛶ fullscreen button (visible only when open) that requests fullscreen on the first `<iframe>` inside |

```mdx
{/* Simple dropdown — no link */}

<Details title="Show Strategy Board">
  ![Phase 1 Strategy Board](/images/ultimates/kefka/p1-strategyboard.webp)
</Details>

{/* Dropdown with external link */}

<Details
  title="Merry-Go-Round — Show Raidplan"
  href="https://docs.google.com/..."
>
  <iframe
    src="https://docs.google.com/..."
    width="100%"
    height="720"
    allowFullScreen
  />
</Details>

{/* Dropdown with external link + fullscreen button (for sites without a built-in fullscreen, e.g. raidplan.io) */}

<Details
  title="Phase 2 — Show Raidplan"
  href="https://raidplan.io/plan/..."
  fullscreen
>
  <iframe
    src="https://raidplan.io/plan/..."
    width="100%"
    height="720"
    allowFullScreen
  />
</Details>
```

---

### `<UnderConstruction>`

A warning banner for pages that are not yet complete.

| Prop                  | Type    | Default                                            | Description                               |
| --------------------- | ------- | -------------------------------------------------- | ----------------------------------------- |
| `title`               | string  | `"Page Under Construction"`                        | Banner heading                            |
| `message`             | string  | `"We're working hard to bring you new content..."` | Body text                                 |
| `showEstimate`        | boolean | `false`                                            | Whether to show an estimated completion   |
| `estimatedCompletion` | string  | `"soon"`                                           | Text shown next to "Expected completion:" |

```mdx
<UnderConstruction />

<UnderConstruction
  title="Phase 4 Guide Coming Soon"
  message="This section is still being written."
  showEstimate
  estimatedCompletion="August 2025"
/>
```

---

## Video Embeds

All video components default to `width="100%"` and render with an `aspect-video` aspect ratio.

### `<YouTube>`

| Prop      | Type   | Default  | Description      |
| --------- | ------ | -------- | ---------------- |
| `videoId` | string | required | YouTube video ID |

```mdx
<YouTube videoId="iBMu1q1dZcA" />
```

---

### `<TwitchClip>`

| Prop      | Type   | Default  | Description    |
| --------- | ------ | -------- | -------------- |
| `videoId` | string | required | Twitch clip ID |

```mdx
<TwitchClip videoId="SomeClipSlug" />
```

---

### `<TwitchVoD>`

Uses the Twitch Embed SDK. Loads lazily.

| Prop      | Type   | Default  | Description   |
| --------- | ------ | -------- | ------------- |
| `videoId` | string | required | Twitch VOD ID |

```mdx
<TwitchVoD videoId="123456789" />
```

---

### `<Streamable>`

| Prop      | Type   | Default  | Description         |
| --------- | ------ | -------- | ------------------- |
| `videoId` | string | required | Streamable video ID |

```mdx
<Streamable videoId="abc123" />
```

---

## Buff Components

Used in detailed mechanic breakdowns. Require a fight-specific TOML data file located alongside the MDX.

### `<Buff>`

Renders an inline buff icon with optional duration and stack count. Hovering shows a tooltip with the buff's name and description.

| Prop       | Type            | Default  | Description                                                              |
| ---------- | --------------- | -------- | ------------------------------------------------------------------------ |
| `b`        | string          | required | Key into the fight's TOML buff data file                                 |
| `dur`      | number / string | —        | Duration text shown below the icon. A bare number is suffixed with `"s"` |
| `short`    | boolean         | `false`  | Omit the buff name label                                                 |
| `stacks`   | number          | —        | Stack count; offsets the icon ID accordingly                             |
| `datapath` | string          | —        | Override path to the TOML data file (rarely needed)                      |

```mdx
<Buff b="Thunderstruck" dur={18} />
<Buff b="Magic Vulnerability Up" stacks={2} short />
```

### `<BuffAppendix>`

Renders a reference table of all buffs defined in the fight's TOML data file. Usually placed at the bottom of a guide.

```mdx
<BuffAppendix />
```
