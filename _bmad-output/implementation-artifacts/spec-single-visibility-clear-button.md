---
title: 'Keep one Visibility formula removal button'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
route: 'one-shot'
---

# Keep one Visibility formula removal button

## Intent

**Problem:** Visibility shows both Clear and Set null buttons, which both remove the condition's effect.

**Approach:** Keep the existing × Clear action as the single formula removal button and remove the Visibility-specific ∅ action.

## Suggested Review Order

- Visibility offers Clear while Background retains its distinct null action.
  [App.tsx:3961](/Users/panitw/Projects/folio/folio-designer/src/App.tsx:3961)

- The existing null-state fixture expects Clear without a second Visibility action.
  [App.test.tsx:9988](/Users/panitw/Projects/folio/folio-designer/src/App.test.tsx:9988)

- The control census matches the reduced button set across selection states.
  [control-vocabulary-contract.test.tsx:434](/Users/panitw/Projects/folio/folio-designer/src/control-vocabulary-contract.test.tsx:434)

- The Line panel's accessibility expectations omit the removed action.
  [line-rect-vocabulary.test.tsx:694](/Users/panitw/Projects/folio/folio-designer/src/line-rect-vocabulary.test.tsx:694)
