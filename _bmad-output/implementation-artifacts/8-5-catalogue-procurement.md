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

| # | Family (declared in the `name` table) | Upstream | Licence | Pinned tag | Path in archive | sha256 of the committed face | Bytes | Thai |
|---|---|---|---|---|---|---|---|---|
| 1 | Inter | rsms/inter | OFL-1.1 | `v4.1` | `extras/ttf/Inter-Regular.ttf` | `40d692fce188e4471e2b3cba937be967878f631ad3ebbbdcd587687c7ebe0c82` | 411,640 | – |
| 2 | Inter Display | rsms/inter | OFL-1.1 | `v4.1` | `extras/ttf/InterDisplay-Regular.ttf` | `99614bda7ff423aaf470990692dd93613a5971ab4446e4a6d5a83b3d74865074` | 408,972 | – |
| 3 | Source Sans 3 | adobe-fonts/source-sans | OFL-1.1 | `3.052R` | `TTF/SourceSans3-Regular.ttf` | `4644c81b86ec9caaa76b634889968ed3c4f4f52f054855933acc7c2b21e53b0f` | 431,196 | – |
| 4 | Source Serif 4 Display | adobe-fonts/source-serif | OFL-1.1 | `4.005R` | `source-serif-4.005_Desktop/TTF/SourceSerif4Display-Regular.ttf` | `a1788d674713c99e61d4541c85e5cfbc3145c4e746b0185c674a0b15f7313295` | 296,680 | – |
| 5 | Source Code Pro | adobe-fonts/source-code-pro | OFL-1.1 | `2.042R-u` | `TTF/SourceCodePro-Regular.ttf` | `74bd80d3e42a08517cd7e1108ba3d86f2da29ac0f3065be95e0357956ab9db37` | 210,312 | – |
| 6 | Fira Code | tonsky/FiraCode | OFL-1.1 | `6.2` | `ttf/FiraCode-Regular.ttf` | `5992ab9640e2df491b2f609467b1de60e8bc39b2c28db184342a0592d98f6117` | 289,624 | – |
| 7 | JetBrains Mono | JetBrains/JetBrainsMono | OFL-1.1 | `v2.304` | `fonts/ttf/JetBrainsMono-Regular.ttf` | `a0bf60ef0f83c5ed4d7a75d45838548b1f6873372dfac88f71804491898d138f` | 273,900 | – |
| 8 | Roboto | googlefonts/roboto-3-classic | OFL-1.1 | `v3.016` | `android/static/Roboto-Regular.ttf` | `e688a215e0841b6e4edb1207c93f88f4c609f82e870884349a7257e449eb9355` | 355,956 | – |
| 9 | Noto Serif | notofonts/latin-greek-cyrillic | OFL-1.1 | `NotoSerif-v2.015` | `NotoSerif/unhinted/ttf/NotoSerif-Regular.ttf` | `a15cfbbc1539d707115111d672d590a3d70d4f74b4c0a315956da20ae19a14e1` | 482,540 | – |
| 10 | Noto Serif Thai | notofonts/thai | OFL-1.1 | `NotoSerifThai-v2.002` | `NotoSerifThai/unhinted/ttf/NotoSerifThai-Regular.ttf` | `538df2b3033522cd48bff18d536f76db60ec04e8cb3f44028fa92dbee8e6e125` | 26,496 | **yes** |
| 11 | Noto Sans Thai Looped | notofonts/thai | OFL-1.1 | `NotoSansThaiLooped-v2.000` | `NotoSansThaiLooped/unhinted/ttf/NotoSansThaiLooped-Regular.ttf` | `6d35234390e9072574f470309bee5d7e0ca9b4d4435803f46085a00fa00ce01a` | 30,976 | **yes** |
| 12 | Cascadia Code | microsoft/cascadia-code | OFL-1.1 | `v2407.24` | `ttf/static/CascadiaCode-Regular.ttf` | `c33ef522cdfeff99907fb54f3e97152bb18bfb9b56ec2fef4d4ceec51c8974a4` | 598,060 | – |
| 13 | Cascadia Mono | microsoft/cascadia-code | OFL-1.1 | `v2407.24` | `ttf/static/CascadiaMono-Regular.ttf` | `06520d032ec274fa5040b22c6f4a1d829081b24ba40b2da56dae89bf10c7b481` | 575,912 | – |
| 14 | **Ubuntu Sans** | canonical/Ubuntu-Sans-fonts | **Ubuntu-font-1.0** | `v1.006` | `UbuntuSans-fonts-1.006/ttf/UbuntuSans-Regular.ttf` | `74f238be44ac5e2ad41021f0b4acc5ccc66f585d06c36b22931319d9751d50ea` | 487,492 | – |
| 15 | **Ubuntu Sans Mono** | canonical/Ubuntu-Sans-Mono-fonts | **Ubuntu-font-1.0** | `v1.100` | `Ubuntu-Sans-Mono-fonts-v1.100/fonts/ttf/UbuntuSansMono-Regular.ttf` | `f2cf9d80e5f4dc4fff6b42ba993dfb3ddcb8ebaf358ef516b9c36a6fd407e04b` | 216,440 | – |
| 16 | Space Grotesk | floriankarsten/space-grotesk | OFL-1.1 | `2.0.0` | `SpaceGrotesk-2.0.0/ttf/static/SpaceGrotesk-Regular.ttf` | `5ede28c4425f3fe4830c8f4754b39e9a87a93d0c3baa5e0a9924532aaa8a98bd` | 114,428 | – |
| 17 | Intel One Mono | intel/intel-one-mono | OFL-1.1 | `V1.4.0` | `ttf/IntelOneMono-Regular.ttf` | `131ab6a8f6e8b9160bc526353262828bc24caab8fcdcfd7a9edc7a044e974230` | 125,928 | – |
| 18 | Geist | vercel/geist-font | OFL-1.1 | `v1.7.2` | `geist-font/Geist/ttf/Geist-Regular.ttf` | `5c8968eafb98a4c4f47033daf29e38e284a6f2a82eb017d171ab040fe7c4b615` | 126,048 | – |
| 19 | Geist Mono | vercel/geist-font | OFL-1.1 | `v1.7.2` | `geist-font/GeistMono/ttf/GeistMono-Regular.ttf` | `42d8ad2e610238e64e8abfcde3037c63f7850a73928742b7ab7229d897bcb155` | 148,516 | – |
| 20 | Literata | googlefonts/Literata | OFL-1.1 | `3.103` | `fonts/ttf/Literata-Regular.ttf` | `0390890de9bb9d5862a6ba4125b82c61792ccc3d66b63e73eee75c1a16fcd208` | 319,964 | – |
| 21 | Cousine | googlefonts/cousine | OFL-1.1 | `v1.241` | `cousine-v1.241/fonts/ttf/Cousine-Regular.ttf` | `1da22250675fc4c42fcf3a9736c44bc0570516105331443b663fd5cfbd1412fe` | 296,856 | – |
| — | IBM Plex Sans | *shipped* | OFL-1.1 | *shipped* | already in tree | *unchanged by this story* | – | – |
| — | IBM Plex Sans Thai | *shipped* | OFL-1.1 | *shipped* | already in tree | *unchanged by this story* | – | **yes** |
| — | IBM Plex Mono | *shipped* | OFL-1.1 | *shipped* | already in tree | *unchanged by this story* | – | – |
| — | Noto Sans | *shipped* | OFL-1.1 | *shipped* | already in tree | *unchanged by this story* | – | – |
| — | Noto Sans Thai | *shipped* | OFL-1.1 | *shipped* | already in tree | *unchanged by this story* | – | **yes** |
| — | Noto Sans SC | *shipped* | OFL-1.1 | *shipped* | already in tree | *unchanged by this story* | – | – |

**PROCURED 2026-09-02.** Every row above was fetched, and the digest recorded is
`shasum -a 256` of the file this repository now commits — not a placeholder.
Each face's own `NOTICE.md` carries the same digest plus the release archive's
digest and byte count, and `folio-designer/src/font-catalogue.test.ts` asserts
the two agree, so a swapped binary reddens rather than quietly falsifying its
own record.

**Three corrections this table takes from the fetch**, recorded rather than
smoothed over:

1. **Ubuntu Sans Mono's in-archive path was wrong as planned.** It was written
   `UbuntuSans-fonts-1.100/fonts/ttf/…`; the archive's real root is
   `Ubuntu-Sans-Mono-fonts-v1.100/`. Corrected above against a real listing.
2. **Five release archives ship no licence file at all** — Source Sans 3,
   Source Code Pro, Fira Code, Roboto and both Cascadia faces. For those the
   licence text is the project's own `LICENSE`/`LICENSE.md`/`OFL.txt` **at the
   pinned tag**, fetched from the same commit the release was cut from. Each
   affected `NOTICE.md` says so in as many words rather than implying the text
   came out of the archive.
3. **The committed licence filenames are all `LICENSE*`**, because
   `manifest.ResolveAssets` matches on that exact uppercase prefix. Canonical's
   file is upstream-named `LICENCE.txt` and several others are `OFL.txt`; the
   TEXT is unmodified upstream and only the filename is normalised. The two
   Ubuntu faces therefore commit `LICENSE-UFL.txt`.

**Source Serif 4's declared family is `Source Serif 4 Display`**, which is what
the Display cut's own `name` table says. The catalogue names families by what
the bytes declare, never by what the plan called them.

**Total committed catalogue bytes: 6,227,936** across 21 faces. Their combined
Brotli weight in the offline release is recorded by the release manifest itself
(`brotli.catalogue.totalBytes`), which is the figure Story 8.4d inherits.

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
