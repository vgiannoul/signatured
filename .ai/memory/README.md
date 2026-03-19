# Memory Directory

This directory is reserved for storing context and learnings from previous AI agent sessions.

## Purpose

AI agents can use this directory to:
- Record important decisions and their rationale
- Store patterns or conventions discovered during work
- Remember project-specific quirks or gotchas
- Track recurring issues and their solutions

## When to Use

Save memory when:
- You discover a non-obvious solution to a problem
- You make an architectural decision that future agents should know
- You identify a pattern that recurs across the codebase
- You encounter a Google API quirk that isn't well-documented

## Format

Create focused memory files by topic:
- `google-api-quirks.md` - Non-standard API behaviors
- `template-patterns.md` - Common template design patterns
- `testing-guidelines.md` - Project-specific test conventions
- `deployment-notes.md` - Production deployment learnings

## What NOT to Store

- Session-specific task details (use plans/ for that)
- Information already in AGENTS.md or README.md
- Temporary context that won't be useful later
- Speculative or unverified information

## Retention

- Keep evergreen knowledge indefinitely
- Update existing memories when better information is found
- Remove outdated or incorrect memories promptly
- Review and consolidate memories quarterly to prevent bloat
