---
id: SPEC-fonts
companions:
  - ./format-changes.md
  - ./font-catalogue.md
  - ../spec-folio/folio-format.md
  - ../../planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md
  - ../../planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Folio — Fonts an Author Can Choose and a File Can Carry

## Why

**A pain to solve, blocking the template author.** The engine ships three Regular-only faces —
Noto Sans, Noto Sans Thai, Noto Sans SC — embedded at build time, and a document's `fonts` map
declares named chains over those face names. Nothing in the designer edits that map, and the
starter document declares exactly one chain, so every template created in the designer offers the
author a font list of length one, called `body`. The author's real question — "can this statement
use our brand's typeface?" — has no answer short of hand-editing the file, and no answer at all
for a face the engine does not ship. Meanwhile the two-user contract the whole product rests on
(the author on one side of the `.folio` file, the Go developer embedding the library on the other)
means any answer has to travel *inside* the file: the developer's server has no font install step,
no network at render time, and no way to be told what was missing. So the work is one shape —
the faces a document uses are declared in it, chosen in the designer, and carried in the file.

## Capabilities

- **CAP-1 — Authorable font chains**
  - **intent:** The author creates, renames and deletes the document's font chains, and reorders
    the entries inside them, so `fontFamily` names something the author chose rather than
    whatever the starter file happened to declare.
  - **success:** A document created from blank in the designer ends up with more than one chain,
    each usable from the typography panel's family control, with no hand-editing of the `.folio`;
    chain and entry edits travel as engine commands and survive a save/open round trip.

- **CAP-2 — Embedded font assets**
  - **intent:** A `.folio` carries the faces its chains name, so a template renders identically on
    a machine that has never seen those faces and has no network.
  - **success:** A document whose chain names an embedded face renders on a clean machine with
    the shipped font set alone available, byte-identically to the designer's own preview; the
    same document saved twice is byte-identical; two elements naming the same face store one copy.

- **CAP-3 — Bundled font catalogue in the designer**
  - **intent:** The author searches a curated catalogue of freely-licensed families and picks one,
    and picking it puts that face in the document.
  - **success:** With the browser offline, the author searches the catalogue, picks a family, and
    the saved file contains that face's bytes and its licence record; no request leaves the
    machine at any point in that flow.

- **CAP-4 — Located failure for a broken font reference**
  - **intent:** A chain entry that names neither a shipped face nor a present, decodable font
    asset fails where it is written, rather than being silently substituted or dropped.
  - **success:** Loading a document whose chain names an absent asset key, or an asset whose bytes
    are not a font this version can read, is a load error naming the chain and the entry; no
    render produces a page set with a substituted face.

## Constraints

- `.folio` stays a **single JSON text file**. No zip, no directory package, no sidecar. The format
  exists so a person or an agent can edit a template without the designer (FR12) and so a
  hand-written template renders (S9); a container ends both, and puts entry order, timestamps and
  compression inside the byte-identity regime.
- Font bytes use the **existing `assets` map** — content-addressed by lowercase hex SHA-256,
  base64 hard-wrapped at 76 columns, deduplicated by key, emitted in stable order. No second
  storage mechanism, no new canonical-serialization rule.
- **Nothing is fetched or read at render time.** No network, no host-installed font, no path on
  disk (FR33). A document renders from its own bytes plus the shipped set.
- **The designer's authoring path is offline too.** The catalogue and its faces ship in the offline
  release bundle behind the service worker's verified asset URLs; no call to `fonts.google.com` or
  `fonts.gstatic.com` at any point.
- **Embedded faces are stored whole.** No save-time subsetting: the data a template renders changes
  between saves, so a subset chosen at save time drops glyphs a later render needs. Subsetting
  stays where it already is — once per render, inside the PDF producer.
- **Byte identity is unchanged.** Fonts embedded, the four targets still produce identical bytes,
  and a no-op round trip still rewrites the file byte-for-byte.
- **Every embedded face carries a licence record**, so a file that travels alone states what it
  redistributes (AD-26).
- **Chain semantics are unchanged**: an ordered list resolved per rune for coverage. An entry may
  name a shipped face or an embedded asset; a chain may mix them.

## Non-goals

- **No container format.** The zip-of-folders shape is explicitly out; revisit only if embedded
  document weight becomes the measured problem, with the CJK case as its trigger.
- **No live font service.** No Google Fonts API, no arbitrary URL, no "download on first use".
- **No host fonts.** Faces installed on the authoring or rendering machine are never enumerated or
  read.
- **No synthetic bold or oblique**, and no variable-font axes. A weight is a face or it does not
  exist.
- **No save-time subsetting**, and no change to how the PDF producer subsets.
- **No CJK families in the embeddable catalogue** in this scope — CJK stays on the shipped-face
  path (a full SC face is 10.6 MB against 646 KB for Latin and 47 KB for Thai).

## Success signal

An author opens the designer offline, picks a brand typeface from the catalogue, lays out the
statement, and saves. The developer's service — a clean Linux box with no fonts installed and no
network — renders that file and gets the same PDF the designer previewed, hash for hash. Nobody
installed a font, and nobody had to be told which one to install.

## Assumptions

- The catalogue is restricted to OFL-licensed families, reading "fonts that can be used freely" as
  redistribution-permitting rather than merely free of charge.
- The new fields are additive, so a MINOR `.folio` version bump; recorded as an open question
  because the version rule is a load-time error path. **SETTLED at Story 8.3: MAJOR, and the
  document declares `2.0`.** The `font` record and the font asset are indeed additive, but the
  chain ENTRY is not: `{"asset": …}` changes the legal shape of an existing value, and a `1.x`
  reader decodes a chain entry as a string and never coerces, so it refuses the file. See Open
  Questions below.

## Open Questions

- ~~MINOR or MAJOR `.folio` version bump for the additive font-asset fields?~~ **SETTLED (Story
  8.3): MAJOR — a document with an embedded-face chain entry declares `2.0`, joining the `2.0`
  Story 7.3 already opened rather than opening a `3.0`.** Derived from three shipped decisions:
  **D-R7.9** (owner) names this story and this version literally — "Epic 8's format change joins the
  same 2.0 at Story 8.3" — and dissolves the tag-deadline objection against it ("we haven't released
  anything to production yet"); **D-7.3.1**'s pre-reader test asks whether a pre-`2.0` reader would
  refuse the file or render it wrong, and a `1.x` reader refuses an object chain entry outright, so
  anything below `2.0` would be a version that lies; and **D-1.4.13** makes `version` a property of
  the DOCUMENT, which is why the trigger is the entry shape and not the presence of a font asset —
  a font asset no chain references loads and renders correctly on a `1.x` reader and must not raise
  the document. `SupportedMajor` does not move: `2.0` is already the library ceiling, and this is a
  second reason a document declares it, not a new rank.
- ~~Which families make the shipped catalogue, how many, and who curates the list as it changes?~~
  **SETTLED (Story 8.5, D-8.5.3 — OWNER DECISION): 20+ families, admitted by a named permissive
  allowlist — OFL-1.1, Apache-2.0, MIT, UFL — enforced the way AD-26's dependency ban is: fail the
  build, never warn, with an unclassifiable licence treated as failure (D-8.5.2).** Three things
  this settles that the bare question does not. **Who curates:** the allowlist does, mechanically —
  admission is a build-time check against the four identifiers, not a standing editorial judgement,
  so the list changes by a licence classification changing and not by anyone's taste. **Why it was
  the owner's call and not engineering's:** AD-26's Rule has two clauses and its heading is not its
  Rule — the family ban is scoped to *"No dependency"*, while redistributed **assets** owe only their
  licence text and copyright lines. Extending the ban to assets was therefore a decision, not a bug
  fix. It is a product and legal call because **fonts do not link**: Folio embeds and subsets a font
  program into the **user's PDF**, so an asset's licence attaches to the documents users produce, not
  to Folio's binary, which is what AD-26's stated Prevents is about. **How many:** 20+ was chosen
  **against** the engineering recommendation of a 5–10 family MVP set, with the 64-asset ceiling and
  the resulting overage stated in the question rather than discovered afterwards — the brief is that
  the catalogue should be big and the engineering should find a shape where size is not paid for at
  first load.
- Do bold and italic get realized in this scope? They are stored and projected today but consumed
  by no producer, and no bold or italic face ships — this is the work that could give them
  meaning, or they stay explicitly out and the panel says so.
- ~~May an author embed a font file from their own disk, the way an image is imported, or is the
  catalogue the only source?~~ **SETTLED (Story 8.6, D-8.6.1): declined — the catalogue is the
  source, and the decline is STATED at the family control rather than left to be inferred from a
  missing button.** Derived rather than chosen. The catalogue exists because a redistributed face
  must arrive with its terms: `font-catalogue.json` names each face's licence, the committed
  `LICENSE*` beside it is the text, and the binary's own `name` table carries the copyright, so the
  designer can fill the `licence`/`licenceText`/`copyright` a `2.0` document now REQUIRES of an
  embedded face (below, and `folio-format.md`). **A file off the author's disk supplies none of
  that** — the designer would have to either invent the terms, ask the author to type them, or
  write a document its own engine refuses to load. The first two are worse answers than declining;
  the third is not available. AD-26's reasoning points the same way: fonts do not link, so an
  embedded face's licence attaches to the documents users produce, and admission is a build-time
  check against a named allowlist (D-8.5.2/D-8.5.3) rather than a judgement made at authoring time
  about a file nobody has seen. The decline is explicit because a control that is simply absent
  reads as an omission; this one is a decision.
- ~~Does the licence record live inline on each font asset, or in one document-level notice
  block?~~ **SETTLED (Story 8.6, D-8.6.1): INLINE, on each font asset, and with the ACTUAL TEXT —
  and, on an asset a chain names, REQUIRED rather than optional.** Derived from what a `.folio` is
  for. The format's whole premise is that the file travels alone (CAP-2: "a `.folio` carries the
  faces its chains name"), and an asset lifted out of one document and put into another must still
  say on what terms it may be passed on — a document-level notice block would not survive that
  move, and a block naming three families would have to be re-partitioned by hand the moment one of
  them is removed. Inline costs duplication (three OFL families in one document carry three
  near-identical copies of ~4 KB), and that is accepted deliberately as the price of the asset
  being self-describing. `licenceText` carries the text and not merely the identifier because
  "OFL-1.1" is a name, and a name is not a grant. **Required, by owner ruling of 2026-09-02:**
  making it mandatory would normally strand earlier documents, and there are none — Folio is
  unreleased, breaking is free now and expensive after `folio-go/v0.1.0` (Story 15.3), so the
  format is made right here. The requirement is scoped to an asset a chain actually NAMES, which is
  what keeps D-1.4.13 intact: an unreferenced font asset is not an embedded face, still loads clean
  with no record at all, and still leaves the document at `1.x`.
- Is an embedded CJK face refused, warned about, or simply allowed at its full weight?
