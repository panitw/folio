# Epic 16 Context: A font arrives from the web, and stays on this machine

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

An author reaches the font library, not a shortlist. Epic 8 left a curated 21-family catalogue chosen by
a build-time licence check and changeable only by releasing the designer; this epic replaces it with a
browser over the whole published library, fetching a family's bytes, its upstream licence text and its
metadata at authoring time, keeping what was fetched on the machine, and showing the author which of
three relationships each font has to their file. What is live is narrower than "live", and the epic is
honest about it: the family index cannot be read from a browser and is a build-time snapshot that ages
between releases, while bytes and terms are fetched on the author's pick. The render boundary does not
move — nothing is fetched at render time, ever.

## Stories

The epics file describes seven stories. Story 16.5 was added later by owner decision, exists in the
sprint tracker only, and has no spec file yet — its title and sequencing are all that is known. Numeric
order is not build order in this epic. Actual order and state:

- Story 16.0: The embed boundary stops throwing, and refuses what render will refuse — done
- Story 16.1b: The binary is asked what it says about itself — done (sequenced before 16.1)
- Story 16.1: A font arrives from the web with its terms attached — done
- Story 16.1a: The local face tier covers the head of the library — done
- Story 16.2: A fetched face stays on this machine — done
- Story 16.3: The font browser is the dialog the design drew — done
- Story 16.5: Installing a face is not embedding it — backlog (runs between 16.3 and 16.4)
- Story 16.4: The family control names three sources — backlog

## Requirements & Constraints

- **Nothing is fetched at render time.** Fetching happens only while authoring. Font resolution inside a
  render stays a pure lookup against an explicit font set, never a host query.
- **Every embedded face carries three records** — a licence identifier, the family's upstream licence
  text, and a copyright read from the face's own name table. Missing any one is a refusal to embed.
- **The licence admission is a run-time decision now, not a build gate.** A closed four-identifier
  allowlist is applied before any byte is embedded; an unclassifiable licence is a refusal, never a
  warning. Nothing copyleft may enter under any declaration.
- **A weight is a static face or it does not exist.** Variable-only families are refused, visibly and
  with their reason; instancing in the browser is refused outright, since it would make the embedded
  face a function of the author's runtime. A family that is variable upstream but ships here as a
  derived static face is offered normally.
- **Only full static TTFs from the one allowed content host.** The CSS/`woff2` route is unusable twice
  over — the engine refuses `woff2` by design, and `unicode-range` subsets would embed partial coverage
  into a document claiming the whole family.
- **The forbidden-host scan is amended, not deleted.** The previously forbidden hosts stay forbidden;
  newly allowed hosts are declared in exactly one module and an occurrence anywhere else still fails the
  build. The population floor and positive control that keep the scan from passing vacuously extend to
  the new half.
- **The index's staleness is stated, not implied** — counts shown are the snapshot's, in words that do
  not claim to be live.
- **Offline degrades honestly**: say so in the product's own terms, offer what the machine already
  holds, and never degrade to a document that will not render.
- **Accessibility floor** on every new surface: keyboard reachable and operable, visible focus,
  accessible names on icon-only controls, state announced. Voice is terse and technical — state the
  fact, name the location, offer no comfort.

## Technical Decisions

- The engine owns the document. There is no TypeScript model of a `.folio`; the UI holds an immutable
  snapshot and sends each committed mutation as one command with one undo. One wasm instance, one
  worker, all calls async.
- Fonts reach the engine as an explicit value; no internal package embeds font data, and font binaries
  live in exactly one place inside the Go module.
- Assets in a `.folio` are keyed by lower-case hex SHA-256 of their bytes. The machine-local face store
  uses that same content address, so store and document agree by construction.
- Persistence is **IndexedDB**, not `localStorage` — a ~5 MB per-origin string quota against a measured
  ~90 KB face that base64 inflates by a third, with a synchronous quota error and no partial write.
- "Available locally" means **faces this machine has fetched**. No operating-system font is enumerated
  or read.
- Guard both doors: the embed command refuses what the renderer refuses, and the renderer keeps its own
  guard — a `.folio` can be hand-written and the loader is not the only entrance.
- The Go licence table is the authority for documents and strictly subsumes the build-time TS one; a
  test fails on divergence rather than a comment warning about it.
- The runtime licence guard reads the binary's own name table and compares it with what is declared: a
  contradiction refuses (located, naming both), a confirmation admits, and **no evidence admits** — the
  threat is a false statement about whose a face is, not a missing one. Any such guard needs a
  positive-control fixture proving it fires.
- Boundary failures must name their cause. A catch whose only vocabulary is "invalid response" cannot be
  debugged from a screenshot, which is how this epic's precondition defect arrived.
- Offline is a build artifact: the service worker precaches shell, wasm, dictionary and font assets.
- Browser-boundary claims are unprovable in Go alone — e2e specs here are compile-checked and not
  executed, so a story either runs a real browser or states plainly that it did not.

## UX & Interaction Patterns

Design source: `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/Font Browser.dc.html`.

- **The browser is a modal**: header with family count and search, a left rail carrying PREVIEW, WRITING
  SYSTEM and CATEGORY, results in Row and Grid views, and a footer counting what is staged. Chips and
  sort filter and order the snapshot's own fields; `reset filters` appears exactly when a filter is
  active; an empty result names the query that produced it.
- **Choosing a font is looking at it**: each result carries a specimen set in that family, at the
  author's chosen size, in the author's own preview text — so a family must be registerable for preview
  independently of any document embedding it.
- A Thai sample toggle swaps the specimen sentence for families covering Thai, which carry a
  `Thai + Latin` badge.
- **Staged multi-select**: several families are staged then confirmed together; the footer reports
  progress rather than freezing, and a family that fails is reported by name without abandoning the rest.
- **The family dropdown names three sources** — `IN THIS TEMPLATE`, `ADDED FROM WEB FONTS`,
  `AVAILABLE LOCALLY` — with a filter field at the top and an `Add fonts…` row pinned at the bottom. The
  three must read as three different relationships to a font without explanation, and a font's group
  changes because the author did something, never on its own.
- The status bar states how many fonts have been added and which is selected.

## Cross-Story Dependencies

- **16.0 is the precondition** — picking a family throws today, so nothing else has ground to stand on.
- **16.1b runs before 16.1**: the runtime licence guard must exist before unreviewed bytes start
  arriving. It narrows the exported embed API and sits on the Epic 15 release gate's before-the-tag list.
- **16.1 supplies the index, fetch and licence path** that 16.2, 16.3 and 16.4 build on. **16.1a** adds
  assets only and changes no screen, so the head of the library works with no download at all.
- **16.2's store populates the `AVAILABLE LOCALLY` group** consumed by 16.3 and 16.4.
- **16.5 runs between 16.3 and 16.4.** It separates installing a face from embedding it, which makes
  16.3's confirm action provisional and redefines what 16.4's three groups mean — building 16.4 first
  would build the control twice.
- **16.4 carries two Epic 8 debts**: the deferred listbox `role="presentation"` finding is fixed there
  rather than multiplied across three group headings, and Story 8.6's disk-font decline is re-derived
  against this epic's premise rather than carried forward unread.
