# Image Assets

## Where to put an image

Put a fight's images in a folder matching its tier and slug under `src/markdown/`:

```text
public/images/<tier>/<fight-slug>/
  banner.avif           # the page banner, always this filename
  <name>.avif           # anything else: cheat sheets, markers, timelines, etc.
```

Example: `public/images/ultimate/dsr/banner.avif`

`savage/` and `extreme/` have one extra level, the expansion, since those tiers get a full new set of fights every expansion:

```text
public/images/savage/<expansion>/<fight-slug>/
public/images/extreme/<expansion>/<fight-slug>/
```

Example: `public/images/savage/dawntrail/m1s/banner.avif`

Keep the source PNG/JPG next to the generated `.avif`, don't delete it.

## Rules

- **Kebab-case filenames.** `pantokrator-diagram.avif`, not `Pantokrator_Diagram.png`.
- **Always `.avif`** for the final file. [#453](https://github.com/naurffxiv/naur/issues/453) automates the conversion, so nobody has to do it by hand.
- **Host images locally.** Don't link to images in the separate `naurffxiv/assets` repo, download them in.
- **Site-wide UI assets:** a few live loose at the top of `public/images/` (not inside a tier folder), things like the site header background or a placeholder graphic. Everything only ever used from a `.tsx` component (icons, etc.) belongs in `src/assets/` instead, imported as a module, not `public/`.

> **TODO/WIP:** [CONTENT_GUIDE.md](CONTENT_GUIDE.md) covers this from the non-technical contributor's side (editing via GitHub's web editor, frontmatter, opening a PR). It's drafted but not yet finalized/reviewed.

## Why it's set up this way

<details>
<summary><code>src/assets/</code> vs <code>public/</code></summary>

Anything only ever used from a `.tsx`/`.ts` file should be `import`ed from `src/assets/`, not dropped in `public/`, imported assets get Next's build-time optimization and an unused one actually gets flagged (a dead duplicate of `naur_icon.png` sat unreferenced in `public/images/` for a long time before [#454](https://github.com/naurffxiv/naur/issues/454) caught it). `server-header.avif` and `under-construction-mammet.avif` are exceptions that have to stay in `public/`: one's read as a raw CSS `url()` in Tailwind config, the other's embedded in `.mdx` content evaluated at runtime rather than statically compiled, neither can go through a JS `import`.

</details>
