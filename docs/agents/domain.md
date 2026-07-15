# Domain Docs

How engineering skills should consume this repo's domain documentation when exploring the codebase.

## Before Exploring

- Read root `CONTEXT.md`.
- Read ADRs in `docs/adr/` that touch the area being changed.

If these files do not exist, proceed silently. The domain-modeling skill creates them lazily when terms or decisions are resolved.

## File Structure

This is a single-context repository:

```
/
├── CONTEXT.md
└── docs/adr/
```

## Glossary Vocabulary

Use terms from `CONTEXT.md` in issue titles, refactor proposals, hypotheses, and test names. Do not drift to synonyms the glossary explicitly avoids.

If a required concept is absent from the glossary, reconsider whether it is needed or note the gap for domain modeling.

## ADR Conflicts

If a proposed change contradicts an ADR, surface the conflict explicitly rather than silently overriding it.
