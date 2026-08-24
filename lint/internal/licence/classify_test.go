package licence

import "testing"

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
