#!/usr/bin/env python3
"""Derive folio-go's three shipped STATIC faces from their upstream variable builds.

Story 2.2 / D-2.2.4 (binding). folio-go ships static, Regular-only instances —
not the upstream variable builds — for three reasons, each measured:

  1. `NotoSansSC-VF`'s `wght` axis has default=100. Pinning axes to their
     DEFAULTS (what the render seam used to do) embedded Simplified Chinese as
     *Thin* beside Regular Latin and Regular Thai. See D-000.21.
  2. `textshape@v0.0.15` `subset/execute.go:496-499` copies `OS/2` verbatim into
     the subset and never updates `usWeightClass` — there is no writer for that
     field anywhere in textshape. So pinning `wght=400` at render time would give
     Regular outlines carrying metadata that still claimed Thin: strictly worse.
     `fontTools` DOES set the field when it instances, so instancing here and
     letting textshape copy it through is what makes the value correct.
  3. Reaching the vendor's `PinAxisLocation` requires the identifier `float32`,
     which `folio-go/internal/arch_test.go:54` bans under `internal/` and the
     module root (AD-23). Deleting the seam satisfies the guard rather than
     fighting it; a caller-supplied face that still carries `fvar` is now
     REJECTED at face ingestion with a located diagnostic naming the remedy.

Why this is a build-time-only tool and its OUTPUT is committed, not generated on
demand: generating at build time would make the shipped font a function of the
build environment — a different fontTools produces a different font, which
produces a different PDF. That is AD-22's drift class reintroduced at the asset
layer. The .ttf files this script writes are committed; this script exists so a
third party can REPLAY the derivation and get the same bytes, and so the
regeneration test (folio-go/fontgen_matrix_test.go, //go:build matrix) can prove
they still reproduce.

Why Python and not Go: `lint`'s `absence-source-date-epoch` content check keys on
the literal string "SOURCE_DATE_EPOCH" appearing in any .go file under
`folio-go/` (D-2.1.5). A Go generator would make that tripwire fire on
legitimate work — the guard is keyed on its purpose (AD-7's params-date wiring,
DW-10), and a font generator is not that purpose.

Three flags are load-bearing, and each was found by measurement, not by reading
documentation:

  --update-name-table   Without it the instanced Regular still names itself
                        'Noto Sans SC Thin' / 'NotoSansSC-Thin', and name[6]
                        becomes the PDF's /BaseFont.
  wdth=100              Latin and Thai carry TWO axes (wght AND wdth). Pinning
                        wght alone leaves `fvar` AND `gvar` alive: the faces
                        would LOOK static and still run the float `gvar` path.
                        Pinning both also cuts Latin 40% compressed
                        (377 KB -> 227 KB), because `gvar` was riding along.
                        NotoSansSC-VF has ONE axis and takes no wdth pin.
  SOURCE_DATE_EPOCH     fontTools is otherwise NOT byte-deterministic:
                        `head.modified` is wall-clock, which drags
                        `checkSumAdjustment` with it. Without this every
                        regeneration differs.

Usage:

    python3 tools/fontgen/instance_faces.py --sources DIR [--repo-root DIR]
                                            [--verify-only] [--out DIR]

`--sources DIR` must hold the three upstream variable builds under the exact
filenames in UPSTREAM below. They are NOT committed (20 MB of inputs for 11 MB of
outputs); each entry records the release URL and the sha256 to fetch them by.

Every hash below is asserted in BOTH directions (D-000.21: assert on the produced
thing, never on the thing you asked for):

  * each SOURCE file is hashed before instancing and must equal `src_sha256`;
  * each PRODUCED file is hashed after instancing and must equal `out_sha256`.

The script refuses to compare an operand it did not successfully read, and prints
an explicit "N of N" coverage witness, so a run that silently produced nothing
cannot report success (D-000.9: absence must not read as success).
"""

from __future__ import annotations

import argparse
import hashlib
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

# The pinned toolchain this derivation is recorded against. A different version
# is not necessarily wrong, but it is not a replay of the recorded derivation,
# so it is reported rather than silently accepted.
EXPECTED_PYTHON = "3.12.13"
EXPECTED_FONTTOOLS = "4.63.0"

# fontTools writes head.modified from the wall clock unless this is set.
# 1451606400 == 2016-01-01T00:00:00Z.
SOURCE_DATE_EPOCH = "1451606400"

UPSTREAM = [
    {
        "key": "Noto Sans",
        "dir": "notosans",
        "src": "NotoSans-VF.ttf",
        "out": "NotoSans-Regular.ttf",
        "axes": ["wght=400", "wdth=100"],
        "family": "Noto Sans",
        "src_sha256": "bfb7bb691513f12e734dc346c03a03f784912432d7e3fa8e56efcf906fe86b3d",
        "src_url": (
            "https://github.com/notofonts/latin-greek-cyrillic/releases/download/"
            "NotoSans-v2.015/NotoSans-v2.015.zip"
            "  ->  NotoSans/googlefonts/variable-ttf/NotoSans[wdth,wght].ttf"
        ),
        "out_sha256": "a4c811314da2ade3b4d2f44e80cda21dd2624e50be82dcefa298dc45a3c92d6c",
        "out_bytes": 646160,
    },
    {
        "key": "Noto Sans Thai",
        "dir": "notosansthai",
        "src": "NotoSansThai-VF.ttf",
        "out": "NotoSansThai-Regular.ttf",
        "axes": ["wght=400", "wdth=100"],
        "family": "Noto Sans Thai",
        "src_sha256": "5a1c559bb539583c8a1fd99d1c5b9491e5e14478c9cd2bd0970d5c3096cc9ef8",
        "src_url": (
            "https://github.com/notofonts/thai/releases/download/"
            "NotoSansThai-v2.002/NotoSansThai-v2.002.zip"
            "  ->  NotoSansThai/googlefonts/variable/NotoSansThai[wdth,wght].ttf"
        ),
        "out_sha256": "c94562c15cbff8c9af93042adb1c63981b5deeeba40693ea8d98cd3b33b73caf",
        "out_bytes": 47788,
    },
    {
        # ONE axis (wght only) — no wdth pin. This is the face whose wght
        # default is 100, i.e. the one D-000.21 was written about.
        "key": "Noto Sans SC",
        "dir": "notosanssc",
        "src": "NotoSansSC-VF.ttf",
        "out": "NotoSansSC-Regular.ttf",
        "axes": ["wght=400"],
        "family": "Noto Sans SC",
        "src_sha256": "a3041811a78c361b1de50f953c805e0244951c21c5bd412f7232ef0d899af0da",
        "src_url": (
            "https://github.com/notofonts/noto-cjk @ "
            "523d033d6cb47f4a80c58a35753646f5c3608a78"
            "  ->  NotoSansSC[wght].ttf (googlefonts variable build)"
        ),
        "out_sha256": "5ef5755b1ac6502180985a2aba6ef0b42d6663829f6f653898ced8411a060158",
        "out_bytes": 10595932,
    },
]


def sha256_of(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as fh:
        for chunk in iter(lambda: fh.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def invocation(face: dict, src: Path, out: Path) -> list[str]:
    """The exact argv this script runs. Reproduced verbatim in each NOTICE.md."""
    return [
        sys.executable,
        "-m",
        "fontTools.varLib.instancer",
        "--update-name-table",
        str(src),
        *face["axes"],
        "-o",
        str(out),
    ]


def shell_form(face: dict) -> str:
    """Human-replayable form of the invocation, for the NOTICE files."""
    return (
        f"SOURCE_DATE_EPOCH={SOURCE_DATE_EPOCH} fonttools varLib.instancer "
        f"--update-name-table \\\n    {face['src']} "
        + " ".join(face["axes"])
        + f" -o {face['out']}"
    )


def check_toolchain() -> tuple[str, str]:
    pyver = ".".join(str(n) for n in sys.version_info[:3])
    try:
        import fontTools  # noqa: PLC0415
    except ImportError:
        sys.exit(
            "fontgen: fontTools is not importable by this interpreter.\n"
            f"  interpreter: {sys.executable}\n"
            f"  install with: {sys.executable} -m pip install "
            f"'fonttools=={EXPECTED_FONTTOOLS}'"
        )
    ftver = fontTools.version
    if pyver != EXPECTED_PYTHON or ftver != EXPECTED_FONTTOOLS:
        print(
            f"fontgen: WARNING — toolchain differs from the recorded derivation.\n"
            f"  recorded: Python {EXPECTED_PYTHON} / fontTools {EXPECTED_FONTTOOLS}\n"
            f"  running:  Python {pyver} / fontTools {ftver}\n"
            f"  The produced-file sha256 assertions below will say whether it "
            f"matters.",
            file=sys.stderr,
        )
    return pyver, ftver


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument(
        "--sources",
        required=True,
        type=Path,
        help="directory holding the three upstream variable builds",
    )
    ap.add_argument(
        "--repo-root",
        type=Path,
        default=Path(__file__).resolve().parents[2],
        help="repository root (default: two levels above this script)",
    )
    ap.add_argument(
        "--out",
        type=Path,
        default=None,
        help="write produced faces here instead of folio-go/fonts/<dir>/ "
        "(used by the regeneration test, which must not touch the tree)",
    )
    ap.add_argument(
        "--verify-only",
        action="store_true",
        help="do not write into the repo; only assert the committed faces "
        "reproduce byte-for-byte",
    )
    args = ap.parse_args()

    pyver, ftver = check_toolchain()

    sources: Path = args.sources
    if not sources.is_dir():
        sys.exit(f"fontgen: --sources {sources} is not a directory")

    print(f"fontgen: Python {pyver} / fontTools {ftver}")
    print(f"fontgen: SOURCE_DATE_EPOCH={SOURCE_DATE_EPOCH}")
    print(f"fontgen: sources {sources}")

    env = dict(os.environ)
    env["SOURCE_DATE_EPOCH"] = SOURCE_DATE_EPOCH

    tmpdir = Path(tempfile.mkdtemp(prefix="fontgen-"))
    compared = 0
    failures: list[str] = []
    try:
        for face in UPSTREAM:
            src = sources / face["src"]
            print(f"\n--- {face['key']} ---")
            if not src.is_file():
                failures.append(
                    f"{face['key']}: source {src} is missing.\n"
                    f"    fetch: {face['src_url']}\n"
                    f"    sha256: {face['src_sha256']}"
                )
                print(f"  SOURCE MISSING: {src}")
                continue

            got_src = sha256_of(src)
            print(f"  source sha256  {got_src}")
            if got_src != face["src_sha256"]:
                failures.append(
                    f"{face['key']}: source sha256 mismatch\n"
                    f"    want {face['src_sha256']}\n"
                    f"    got  {got_src}"
                )
                print("  SOURCE HASH MISMATCH — not instancing this face")
                continue

            produced = tmpdir / face["out"]
            argv = invocation(face, src, produced)
            print(f"  $ SOURCE_DATE_EPOCH={SOURCE_DATE_EPOCH} " + " ".join(argv))
            proc = subprocess.run(
                argv, env=env, capture_output=True, text=True, check=False
            )
            if proc.returncode != 0:
                failures.append(
                    f"{face['key']}: instancer exited {proc.returncode}\n"
                    f"{proc.stderr.strip()}"
                )
                print(f"  INSTANCER FAILED ({proc.returncode})")
                continue
            if not produced.is_file():
                failures.append(f"{face['key']}: instancer wrote no output file")
                print("  NO OUTPUT FILE")
                continue

            got_out = sha256_of(produced)
            got_bytes = produced.stat().st_size
            print(f"  produced       {got_bytes} B  sha256 {got_out}")
            compared += 1
            if got_out != face["out_sha256"] or got_bytes != face["out_bytes"]:
                failures.append(
                    f"{face['key']}: PRODUCED FILE DIFFERS from the recorded "
                    f"derivation\n"
                    f"    want {face['out_bytes']} B  {face['out_sha256']}\n"
                    f"    got  {got_bytes} B  {got_out}"
                )
                print("  MISMATCH vs recorded derivation")
                continue
            print("  matches the recorded derivation")

            dest_dir = (
                args.out
                if args.out is not None
                else args.repo_root / "folio-go" / "fonts" / face["dir"]
            )
            if args.verify_only:
                committed = (
                    args.repo_root / "folio-go" / "fonts" / face["dir"] / face["out"]
                )
                if not committed.is_file():
                    failures.append(
                        f"{face['key']}: committed face {committed} is missing"
                    )
                    print(f"  COMMITTED FACE MISSING: {committed}")
                    continue
                got_committed = sha256_of(committed)
                if got_committed != got_out:
                    failures.append(
                        f"{face['key']}: committed face differs from a fresh "
                        f"derivation\n"
                        f"    committed {got_committed}\n"
                        f"    derived   {got_out}"
                    )
                    print("  COMMITTED FACE DIFFERS")
                else:
                    print(f"  committed face reproduces: {committed}")
            else:
                dest_dir.mkdir(parents=True, exist_ok=True)
                shutil.copy2(produced, dest_dir / face["out"])
                print(f"  wrote {dest_dir / face['out']}")
    finally:
        shutil.rmtree(tmpdir, ignore_errors=True)

    # Coverage witness. A run that compared nothing must not read as success
    # (D-000.9) — my own earlier reproducibility check printed IDENTICAL for
    # four faces that were never generated, because two failed hashes compared
    # empty to empty.
    total = len(UPSTREAM)
    print(f"\nfontgen: derived and compared {compared} of {total} faces")
    if failures:
        print("\nfontgen: FAILED")
        for f in failures:
            print("  * " + f)
        return 1
    if compared != total:
        print(f"\nfontgen: FAILED — coverage witness {compared} != {total}")
        return 1
    print("fontgen: OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
