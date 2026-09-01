This stub module contains NO third-party code, and no Ubuntu-licensed
font. It exists only to be classified by Story 1.3's licence checker
fixture graph (D-1.3.8): a module we authored ourselves, whose LICENSE
file is fixture data we wrote in order to be classified, creates no
licensee relationship and transfers no copyrighted work. It is not a
real software licence grant.

Added at Story 8.4h (D-8.5.3) so `Ubuntu-font-1.0` — the fourth member
of the owner's asset allowlist — is exercised END TO END through the
graph scan, and not only through a classifier table.

WHAT THIS FIXTURE PROVES, EXACTLY. `ScanLicenceGraph` switches on the
licence FAMILY and discards the SPDX id, and the permissive subtest
asserts zero findings — so this fixture proves that `Ubuntu-font-1.0`
resolves to a PERMISSIVE family through the real scan path, which is
to say that the `permissiveSPDX` entry is live rather than declared.
Removing that entry reds this subtest loudly, because the id then
classifies as unknown. It does NOT pin the id itself: any other
permissive identifier here (`MIT`, `ISC`) would also produce a green.
The id is pinned by `TestUbuntuFontLicenceSPDXLineIsPermissive` in
lint/internal/licence/classify_test.go, which asserts it directly.

Nothing in this repository ships under this licence yet; the first real
asset witness arrives with Story 8.5 (Design Note 8).
