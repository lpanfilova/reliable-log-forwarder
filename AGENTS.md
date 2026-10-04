# Project

This repository contains a Go-based telemetry/log forwarding system.

The project is being developed incrementally. Prefer small working changes over large speculative architecture.

## Current priority

Build the simplest working end-to-end pipeline first:

generator -> forwarder -> HTTP receiver

Do not implement future features unless the task explicitly asks for them.

## Engineering guidelines

- Write clear, idiomatic Go.
- Keep functions and components simple.
- Avoid unnecessary abstractions.
- Avoid adding new dependencies unless they are clearly needed.
- Add tests for new behavior.
- Run `gofmt` on changed Go files.
- Run `go test ./...` before completing a task.
- Do not silently ignore errors.
- Do not refactor unrelated code while completing a scoped task.

## Codex task behavior

When given a task:

1. Inspect the relevant existing code first.
2. Implement only the requested scope.
3. Add or update tests.
4. Run formatting and tests.
5. Briefly summarize what changed, important tradeoffs, and any assumptions.
6. Report any tests or commands that failed.

If requirements are ambiguous, prefer the simplest implementation consistent with the current architecture.

## Project goals

The system will eventually be used to study reliability and performance tradeoffs in telemetry delivery.

Future work may include:

- batching
- buffering/backpressure
- retries
- persistent delivery
- metrics
- fault injection
- performance experiments
- agentic incident diagnosis

These are not requirements for the initial MVP.

If the default Go cache is inaccessible, use a temporary cache location outside the repository. Do not create build/cache directories inside the repository.