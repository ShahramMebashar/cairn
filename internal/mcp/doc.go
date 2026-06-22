// Package mcp exposes the 7 agent verbs (list, get, create, claim, transition,
// run_checks, note) as thin adapters over store + task + check (SPEC §7). Gate logic
// is not reimplemented here — verbs call task.Ready / task.CanTransition. Implementation
// lands in build-order step 6.
package mcp
