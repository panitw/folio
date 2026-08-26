package folio

import "github.com/panitw/folio/folio-go/internal/expr"

// Version is the folio-go module version. It is recorded in the golden
// fixture (fixtures/minimal-rect/expected.json) alongside the Go toolchain
// version that produced it. A real release tag is folio-go/v0.1.0
// (directory-prefixed, per AD-22) — that is not this story's business.
const Version = "0.0.0-dev"

// LocaleTableVersion is AC6/AD-22, surfaced here (Finding 9, Story
// 3.4's QA review): internal/expr's locale table carries its own
// version, but being unexported and living under internal/, it could
// not previously reach anywhere the library version is surfaced —
// AD-22 makes it "part of the library version" and D-3.4.3's ruling
// rests on that link existing. Defined as exactly expr.LocaleTableVersion
// (never a second literal) so the two can never drift.
const LocaleTableVersion = expr.LocaleTableVersion
