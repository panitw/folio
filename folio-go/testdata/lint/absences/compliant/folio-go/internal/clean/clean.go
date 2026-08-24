// Package clean is a fixture file for TestAbsencesFixtureScan/compliant
// (Story 2.1's reopening, Finding 13): it exists so the content-kind
// absence check (absence-source-date-epoch) scans a real, non-empty
// folio-go/ subtree here and reports zero findings because there is
// genuinely nothing to find — not because the scope directory is
// absent. See AbsencesStats.ContentFilesScanned, asserted non-zero for
// this fixture leg.
package clean
