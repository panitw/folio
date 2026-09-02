# Story 8.5 — catalogue procurement record

The per-face procurement table for Story 8.5, held here rather than inline in the spec so the spec
stays readable. **Normative for procurement**: a face not on this table, or taken from a different
tag or archive path than the one recorded here, is not procured to standard.

Cited by `8-5-a-curated-catalogue-ships-with-the-designer.md` (AC1, AC3).

## The standard each row must meet

Every catalogue face takes the **vendored static** route (D-8.5.4). This project cannot derive a new
face — `tools/fontgen/instance_faces.py` drives a hardcoded 3-entry `UPSTREAM` list (`:117`–`:168`),
`.font-sources/` is gitignored with zero tracked files, and the bootstrap needs an `out_sha256`
unknowable before the first run. **A face that would require derivation is a Block If, not a task.**

Provenance recorded per face, in a `NOTICE.md` beside the binary, on the shape
`folio-designer/public/fonts/ibmplexsans/NOTICE.md` already sets:

| Field | Meaning |
|---|---|
| Upstream project + release tag | The pinned upstream version. A moving tag is not a pin. |
| Download URL | The release archive actually fetched. |
| Path inside the archive | Which file in the archive became the shipped face. |
| sha256 + byte count of the archive | What was fetched. |
| sha256 of the shipped file | What was committed. Must equal `shasum -a 256` of the binary. |
| Size in bytes | The committed face. |
| Fetched | The date of the fetch. |
| Relation to source | For every row here: *copied unmodified, no derivation*. |

Plus a `LICENSE*` file carrying the unmodified upstream licence text, and a `Copyright` line in the
`NOTICE.md`. The licence gate reads both and fails the build without them.

## Licence constraint

Four identifiers only (D-8.5.3): `OFL-1.1`, `Apache-2.0`, `MIT`, `Ubuntu-font-1.0`. Admission is
**per-term** — a compound expression is admitted only if *every* term is one of the four.

**No `WITH` expression, in any form.** The classifier does not parse `WITH`: it reaches
`unsupported SPDX operator "WITH"` and returns `FamilyUnknown`, so the build fails closed.

> **This is parser scope, not licence policy.** A family excluded here is excluded **pending a
> parser widening**, never because its licence is unacceptable. When someone asks for one of these
> families, the answer is *"widen the parser"*.

The named casualty is **Linux Libertine**, whose actual expression is
`OFL-1.1 OR GPL-2.0-or-later WITH Font-exception-2.0` — a licence that *also* offers plain OFL and
is in every substantive sense acceptable. It is left out because of a comma-level parsing limit.
The capability is registered to Epic 15 (D-8.5.16); it is not lost, only deferred.

## Tier A — procurement pool

Verified upstream-by-upstream: licence identifier read from the project's own `LICENSE` text (not
from a repository-host guess, which reports `NOASSERTION` for several of these), release tag pinned,
and the in-archive path confirmed against a real listing of the archive.

**All rows are static, upright Regular, single-instance, `.ttf`.** No bold, no italic, no oblique,
no variable axes (AC6). No row carries a `WITH` clause; all are single-identifier grants.

Rows marked **shipped** already exist in the tree and are *not* new catalogue additions; they are
listed so the table accounts for the whole bundle.

| # | Family | Upstream | Licence | Pinned tag | Path in archive | Thai |
|---|---|---|---|---|---|---|
| 1 | Inter | rsms/inter | OFL-1.1 | v4.1 | `extras/ttf/Inter-Regular.ttf` | – |
| 2 | Inter Display | rsms/inter | OFL-1.1 | v4.1 | `extras/ttf/InterDisplay-Regular.ttf` | – |
| 3 | Source Sans 3 | adobe-fonts/source-sans | OFL-1.1 | 3.052R | `TTF/SourceSans3-Regular.ttf` | – |
| 4 | Source Serif 4 | adobe-fonts/source-serif | OFL-1.1 | 4.005R | `source-serif-4.005_Desktop/TTF/SourceSerif4Display-Regular.ttf` | – |
| 5 | Source Code Pro | adobe-fonts/source-code-pro | OFL-1.1 | 2.042R-u | `TTF/SourceCodePro-Regular.ttf` | – |
| 6 | Fira Code | tonsky/FiraCode | OFL-1.1 | 6.2 | `ttf/FiraCode-Regular.ttf` | – |
| 7 | JetBrains Mono | JetBrains/JetBrainsMono | OFL-1.1 | v2.304 | `fonts/ttf/JetBrainsMono-Regular.ttf` | – |
| 8 | Roboto | googlefonts/roboto-3-classic | OFL-1.1 | v3.016 | `android/static/Roboto-Regular.ttf` | – |
| 9 | Noto Serif | notofonts/latin-greek-cyrillic | OFL-1.1 | NotoSerif-v2.015 | `NotoSerif/unhinted/ttf/NotoSerif-Regular.ttf` | – |
| 10 | Noto Serif Thai | notofonts/thai | OFL-1.1 | NotoSerifThai-v2.002 | `NotoSerifThai/unhinted/ttf/NotoSerifThai-Regular.ttf` | **yes** |
| 11 | Noto Sans Thai Looped | notofonts/thai | OFL-1.1 | NotoSansThaiLooped-v2.000 | `NotoSansThaiLooped/unhinted/ttf/NotoSansThaiLooped-Regular.ttf` | **yes** |
| 12 | Cascadia Code | microsoft/cascadia-code | OFL-1.1 | v2407.24 | `ttf/static/CascadiaCode-Regular.ttf` | – |
| 13 | Cascadia Mono | microsoft/cascadia-code | OFL-1.1 | v2407.24 | `ttf/static/CascadiaMono-Regular.ttf` | – |
| 14 | **Ubuntu Sans** | canonical/Ubuntu-Sans-fonts | **Ubuntu-font-1.0** | v1.006 | `UbuntuSans-fonts-1.006/ttf/UbuntuSans-Regular.ttf` | – |
| 15 | **Ubuntu Sans Mono** | canonical/Ubuntu-Sans-Mono-fonts | **Ubuntu-font-1.0** | v1.100 | `UbuntuSansMono-fonts-1.100/fonts/ttf/UbuntuSansMono-Regular.ttf` | – |
| 16 | Space Grotesk | floriankarsten/space-grotesk | OFL-1.1 | 2.0.0 | `SpaceGrotesk-2.0.0/ttf/static/SpaceGrotesk-Regular.ttf` | – |
| 17 | Intel One Mono | intel/intel-one-mono | OFL-1.1 | V1.4.0 | `ttf/IntelOneMono-Regular.ttf` | – |
| 18 | Geist | vercel/geist-font | OFL-1.1 | v1.7.2 | `geist-font/Geist/ttf/Geist-Regular.ttf` | – |
| 19 | Geist Mono | vercel/geist-font | OFL-1.1 | v1.7.2 | `geist-font/GeistMono/ttf/GeistMono-Regular.ttf` | – |
| 20 | Literata | googlefonts/Literata | OFL-1.1 | 3.103 | `fonts/ttf/Literata-Regular.ttf` | – |
| 21 | Cousine | googlefonts/cousine | OFL-1.1 | v1.241 | `cousine-v1.241/fonts/ttf/Cousine-Regular.ttf` | – |
| — | IBM Plex Sans | IBM/plex | OFL-1.1 | *shipped* | already in tree | – |
| — | IBM Plex Sans Thai | IBM/plex | OFL-1.1 | *shipped* | already in tree | **yes** |
| — | IBM Plex Mono | IBM/plex | OFL-1.1 | *shipped* | already in tree | – |
| — | Noto Sans | notofonts | OFL-1.1 | *shipped* | already in tree | – |
| — | Noto Sans Thai | notofonts | OFL-1.1 | *shipped* | already in tree | **yes** |
| — | Noto Sans SC | notofonts | OFL-1.1 | *shipped* | already in tree | – |

**Count: 21 new families**, on top of 6 already shipped. AC3 needs ≥20 new, so the pool clears the
bar **by one**. That margin is thin and is stated rather than smoothed over — see the reserve below.

**Two rows are load-bearing beyond their own entry.** `classify.go:167-171` records that
*"NOTHING IN THIS REPOSITORY IS UBUNTU-LICENSED TODAY … there is no analogue of
TestCommittedOFLTextClassifiesAsOFL11 to write until Story 8.5 lands a face under it"*. Rows 14 and
15 are that face. Procuring at least one of them converts the `Ubuntu-font-1.0` allowlist entry from
fixture-only coverage into a real committed asset. **Dropping both silently re-opens that gap.**

## Reserve — draw on these only if a Tier A row fails procurement

Licence confirmed by reading the licence text; the procurement route is weaker (the pinned release
carries no attached archive, so the face comes from the git tag tarball — a different provenance
story, and one the `NOTICE.md` must state honestly).

- **Fira Sans**, **Fira Mono** — mozilla/Fira 4.202, OFL-1.1. Upstream repository is archived.
- **Work Sans** — v2.010, OFL-1.1, `fonts/static/TTF/WorkSans-Regular.ttf`.
- **Merriweather** — SorkinType v2.001, OFL-1.1, `fonts/ttf/Merriweather-Regular.ttf`.
- **Victor Mono** — v1.5.6, OFL-1.1, `src/assets/VictorMono-Regular.ttf`.
- **Red Hat Text**, **Red Hat Display** — v1.0.0, OFL-1.1, `TTF/RedHatText-Regular.ttf`. Later tags
  may be variable-only; pin v1.0.0 or verify before use.
- **IBM Plex Serif**, **Noto Sans Mono** — OFL-1.1, tagged releases exist, **in-archive path was
  inferred from sibling archives and not confirmed**. Verify the path before committing to either.

A second contingency, if the reserve is also exhausted: the `.otf` families **Overpass**,
**Overpass Mono**, **Cormorant**, **Cormorant Infant** and the **Monaspace** set are all OFL-1.1 and
otherwise qualify. They are held back only because the generated `@font-face` emitter hardcodes
`format('truetype')`, so admitting them needs a format branch. That is a bounded, named change —
not a licence question.

## Excluded, with the reason

Recorded because a fail-closed gate makes exclusions load-bearing: someone will otherwise re-propose
these, and two of them would fail the build.

| Family | Reason |
|---|---|
| **Linux Libertine** | `OFL-1.1 OR GPL-2.0-or-later WITH Font-exception-2.0`. Refused for **parser scope, pending a parser widening** — *not* because the licence is unacceptable. |
| **Hack** | Compound: MIT **+ Bitstream Vera** **+ public domain**. Bitstream Vera is outside the four ids, so per-term admission must refuse it. Correctly refused. |
| **Public Sans** | Declared joint work: OFL-1.1 **+ CC0-1.0**. `CC0-1.0` is permissive to the classifier but is **not** on the four-id font allowlist, so per-term admission refuses it. |
| **DejaVu** | Bitstream Vera — outside the allowlist. |
| **Go fonts** | BSD-3-Clause — outside the allowlist. |
| **Sarabun, Kanit, Prompt, Mitr, Chakra Petch, Bai Jamjuree** | All OFL-1.1 and all desirable for Thai, but their upstream repositories carry **no tags**, so there is nothing to pin. Procurement-blocked, not licence-blocked. |
| **Karla, Nunito, Atkinson Hyperlegible, Lexend, Rubik, Arimo, Tinos, Mulish, Playfair, Libre Baskerville, EB Garamond, Lato** | No tagged upstream release to pin. |
| All CJK families | Out of scope (`SPEC.md` `## Non-goals`); a full SC face is 10.6 MB. |

**A consequence worth stating plainly:** every Thai-covering face in the pool comes from just two
vendors, Noto and IBM Plex, because the tagless Cadson Demak families are unreachable under the
pinned-release rule. Thai coverage is therefore adequate but not diverse, and that is a known
limitation of this catalogue rather than an oversight.

## What this table does not settle

Which family the author actually picks, and what happens when they do, is **Story 8.6**. This story
ships the bytes and proves they are there; nothing here makes a family selectable.
