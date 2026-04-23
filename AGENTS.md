# Agents Guide

This repository is being built as both a working application and a portfolio project for a student developer. Agents working here should optimize for two outcomes at the same time:

1. Ship a functioning, maintainable Masters Pool app.
2. Leave the project easier to understand after every change.

That means professional standards matter, but so does explanation. Do not treat this repo like a speedrun where the fastest path is to dump code with no reasoning.

## Project Context

- The current target architecture is documented in `PLAN.md`.
- The current execution checklist is documented in `TODO.MD`.
- `README.md` is still incomplete, so do not assume it is the source of truth yet.
- The intended stack today is Go backend, Vue 3 frontend, PostgreSQL, Clerk auth, and Kubernetes deployment.

Before making major changes, read `PLAN.md` and `TODO.MD` so implementation stays aligned with the planned architecture and current phase.

## Primary Working Principles

### 1. Teach while building

This repo is for learning, not just output. When you make non-trivial changes:

- Explain what you changed in plain language.
- Explain why the approach is appropriate.
- Call out important tradeoffs when they matter.
- Prefer small, understandable steps over dense, clever rewrites.

Do not rely on "just copy this pattern" style work. The student should be able to learn from the change history and your explanations.

### 2. Prefer professional practices

Treat this like a real production-minded project:

- Keep code modular and readable.
- Use clear names.
- Avoid hidden magic and unnecessary abstraction.
- Add or update tests when practical.
- Validate assumptions instead of guessing.
- Keep setup, docs, and environment instructions accurate.

When choosing between two valid approaches, prefer the one that is easier to maintain and explain.

### 3. Keep the app working at each stage

`TODO.MD` is organized in phases. Preserve that idea.

- Do not start broad refactors that leave the repo in a half-broken state without strong reason.
- Try to complete work in vertical slices that leave the app runnable or closer to runnable.
- If a task introduces temporary incompleteness, document it clearly in the final message and update planning docs if needed.

## Required Documentation Maintenance

Agents must keep planning documents current as the project evolves.

### Update `TODO.MD` when:

- A checklist item is completed.
- New work is discovered.
- Scope changes.
- A task should be reordered, split, or removed.

Mark completed items clearly and add newly discovered work in the most appropriate phase. Do not let `TODO.MD` drift away from reality.

### Update `PLAN.md` when:

- The architecture changes.
- A technology choice changes.
- The deployment approach changes.
- The data model or API surface changes in a meaningful way.

`PLAN.md` should describe the intended structure of the system as it actually exists or is now expected to exist.

### Update `README.md` when:

- Setup instructions become known or change.
- Important repo structure becomes clearer.
- A new contributor would otherwise be confused.

If `README.md` remains incomplete, note that explicitly rather than pretending the documentation is finished.

## How to Communicate Changes

For meaningful work, do not stop at "changed X."

Include:

- What changed
- Why it changed
- What the student should learn from it
- Any follow-up work or risks

Use explanations that are technically correct but not overloaded with jargon. Assume the reader is learning professional engineering habits and wants to understand the reasoning, not just the result.

## Implementation Guidance

- Follow the architecture in `PLAN.md` unless there is a clear reason to change it.
- If you intentionally diverge from the plan, update `PLAN.md` in the same piece of work.
- Check whether the work should also update `TODO.MD` before finishing.
- Prefer incremental changes over large speculative scaffolding.
- Keep frontend and backend boundaries clear.
- Keep API contracts explicit.
- Avoid introducing dependencies without a concrete reason.

## Quality Bar

Good work in this repo should usually be:

- Readable by a student revisiting it later
- Structured in a way that would not be embarrassing in a professional code review
- Documented enough that the next step is obvious
- Honest about unfinished pieces

If something is stubbed, say so. If something is a shortcut, label it as a shortcut and explain the proper next step.

## When Requirements Shift

Project requirements will change as learning happens and the app takes shape. Agents are expected to respond to that, not ignore it.

When new requirements, constraints, or better decisions emerge:

- update `PLAN.md` if the intended architecture or design changes
- update `TODO.MD` to reflect the new execution plan
- mention the change explicitly in your summary so the decision is visible

Planning documents are living project artifacts, not one-time files.

## Default Attitude

Be rigorous, practical, and educational.

Build the real app, but leave a trail of explanations strong enough that this repository also works as evidence of learning, decision-making, and professional growth.
