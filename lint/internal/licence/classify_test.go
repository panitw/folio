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
