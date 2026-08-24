package licence

import (
	"os"
	"testing"
)

// TestClassifyLicenceText is AC31's supplement: direct unit tests over
// licence-identification inputs, additional to the fixture graphs
// (AC29), never a substitute for them and never built from a recorded
// `go list -m -json all` output (D-1.3.8).
func TestClassifyLicenceText(t *testing.T) {
	cases := []struct {
		name       string
		text       string
		wantFamily Family
	}{
		{"SPDX MIT", "SPDX-License-Identifier: MIT\n", FamilyPermissive},
		{"SPDX Apache-2.0", "SPDX-License-Identifier: Apache-2.0\n", FamilyPermissive},
		{"SPDX BSD-3-Clause", "SPDX-License-Identifier: BSD-3-Clause\n", FamilyPermissive},
		{"SPDX GPL-3.0-only", "SPDX-License-Identifier: GPL-3.0-only\n", FamilyCopyleft},
		{"SPDX LGPL-3.0-only", "SPDX-License-Identifier: LGPL-3.0-only\n", FamilyCopyleft},
		{"SPDX AGPL-3.0-only", "SPDX-License-Identifier: AGPL-3.0-only\n", FamilyCopyleft},
		{"SPDX SSPL-1.0", "SPDX-License-Identifier: SSPL-1.0\n", FamilyCopyleft},
		{"SPDX CC0-1.0", "SPDX-License-Identifier: CC0-1.0\n", FamilyPermissive},                                  // AC8, D-2.1.3
		{"full CC0 legal code text marker", "Creative Commons Legal Code\n\nCC0 1.0 Universal", FamilyPermissive}, // AC8, D-2.1.3 — matches the committed LICENSE-CC0-1.0.txt shape
		// Red-proof for this story's second QA review, Major 1: the CC0
		// fallback marker must NOT match the boilerplate disclaimer
		// shared by every Creative Commons legal code. A NonCommercial
		// or ShareAlike licence classifying as permissive CC0-1.0 is a
		// silent wrong answer, not a loud miss (D-1.3.8, D-2.1.3). Before
		// this fix these three cases all classified FamilyPermissive/"CC0-1.0".
		{"CC BY-NC-SA legal code must NOT classify as CC0", "Creative Commons Legal Code\n\nCC BY-NC-SA 4.0\n\nCreative Commons Corporation is not a law firm and does not provide legal services or legal advice.", FamilyUnknown},
		{"CC BY-SA legal code must NOT classify as CC0", "Creative Commons Legal Code\n\nAttribution-ShareAlike 3.0\n\nCreative Commons Corporation is not a law firm and does not provide legal services or legal advice.", FamilyUnknown},
		{"CC BY-ND legal code must NOT classify as CC0", "Creative Commons Legal Code\n\nAttribution-NoDerivatives 4.0\n\nCreative Commons Corporation is not a law firm and does not provide legal services or legal advice.", FamilyUnknown},
		{"full GPL text marker", "GNU GENERAL PUBLIC LICENSE\nVersion 3", FamilyCopyleft},
		{"full AGPL text marker", "GNU AFFERO GENERAL PUBLIC LICENSE", FamilyCopyleft},
		{"full SSPL text marker", "Server Side Public License, Version 1", FamilyCopyleft},
		{"full MIT text marker", "MIT License\n\nPermission is hereby granted, free of charge", FamilyPermissive},
		{"empty", "", FamilyUnknown},
		{"unrecognised marker", "Some Proprietary Thing v2", FamilyUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotFamily, _ := ClassifyLicenceText(c.text)
			if gotFamily != c.wantFamily {
				t.Errorf("ClassifyLicenceText(%q) family = %v, want %v", c.text, gotFamily, c.wantFamily)
			}
		})
	}
}

// TestClassifyOFL covers the OFL branch, which shipped with no test at
// all — positive or negative — while the CC0 marker immediately below it
// carries three negative red-proofs added after an earlier QA finding on
// exactly this over-match class.
//
// The FAMILY is not the interesting assertion here: OFL is permissive
// whatever version it is, so the copyleft gate was never at risk. The
// SPDX ID is what lands in lint/MANIFEST.md and attributes each shipped
// face, so the id is asserted explicitly.
func TestClassifyOFL(t *testing.T) {
	// The real committed text's opening lines, verbatim in shape.
	const oflHeader = "Copyright 2022 The Noto Project Authors\n\n" +
		"This Font Software is licensed under the SIL Open Font License, Version 1.1.\n" +
		"This license is copied below, and is also available with a FAQ at:\n" +
		"https://openfontlicense.org\n\n" +
		"SIL OPEN FONT LICENSE\nVersion 1.1 - 26 February 2007\n"

	cases := []struct {
		name       string
		text       string
		wantFamily Family
		wantID     string
	}{
		{
			"OFL 1.1 full text classifies as OFL-1.1",
			oflHeader,
			FamilyPermissive, "OFL-1.1",
		},
		{
			// The defect the OFL branch exists to fix: OFL 1.1's grant
			// clause opens with MIT's detection phrase verbatim, so
			// before the branch was added ABOVE the MIT case, every
			// shipped face classified as "MIT".
			"OFL text carrying MIT's trigger phrase must NOT classify as MIT",
			oflHeader + "\nPermission is hereby granted, free of charge, to any person obtaining\n" +
				"a copy of the Font Software, to use, study, copy, merge, embed, modify,\n" +
				"redistribute, and sell modified and unmodified copies of the Font Software.\n",
			FamilyPermissive, "OFL-1.1",
		},
		{
			// Over-match negative. OFL 1.0's text contains the identical
			// licence name, so a name-only match labelled it OFL-1.1 —
			// a silent wrong answer. FamilyUnknown is a LOUD failure
			// (D-1.3.4 makes it a build failure), which is the correct
			// outcome for a licence this classifier does not know.
			"OFL 1.0 must NOT be labelled OFL-1.1",
			"SIL OPEN FONT LICENSE\nVersion 1.0 - 22 November 2005\n",
			FamilyUnknown, "",
		},
		{
			// A dependency LICENSE that merely BUNDLES OFL text would
			// otherwise outrank MIT/Apache/BSD purely by ordering.
			"a bundled OFL 1.0 notice does not win over the file's own licence",
			"MIT License\n\nPermission is hereby granted, free of charge\n\n" +
				"Bundled fonts are under the SIL Open Font License.\n",
			FamilyPermissive, "MIT",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotFamily, gotID := ClassifyLicenceText(c.text)
			if gotFamily != c.wantFamily {
				t.Errorf("family = %v, want %v", gotFamily, c.wantFamily)
			}
			if gotID != c.wantID {
				t.Errorf("SPDX id = %q, want %q", gotID, c.wantID)
			}
		})
	}
}

// TestCommittedOFLTextClassifiesAsOFL11 reads the ACTUAL committed
// licence file rather than a hand-written approximation of it (D-000.21:
// assert on the artifact). A synthetic fixture that drifts from the real
// text would keep passing while the real classification broke.
func TestCommittedOFLTextClassifiesAsOFL11(t *testing.T) {
	const rel = "../../../folio-go/fonts/notosans/LICENSE-OFL.txt"
	data, err := os.ReadFile(rel)
	if err != nil {
		t.Fatalf("read committed OFL text %s: %v", rel, err)
	}
	if len(data) == 0 {
		t.Fatalf("%s is empty — this assertion would be vacuous", rel)
	}
	family, id := ClassifyLicenceText(string(data))
	if family != FamilyPermissive || id != "OFL-1.1" {
		t.Fatalf(
			"the committed OFL text classifies as (%v, %q), want (FamilyPermissive, %q) — "+
				"this is the text that attributes all three shipped faces in lint/MANIFEST.md",
			family, id, "OFL-1.1",
		)
	}
}
