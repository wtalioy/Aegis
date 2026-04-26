# Backend Test Layout

- `tests/unit`: pure service and domain behavior
- `tests/integration`: in-process runtime and wiring tests
- `tests/contract`: HTTP/API contract tests
- `tests/fakes`: reusable test doubles
- `tests/helpers`: shared builders and assertions

## Naming Conventions

- Prefer file names in the form `<subject>_<scope>_test.go` when a suite owns a clear behavior area.
- Prefer test names in the form `Test<Subsystem>_<Behavior>` so failures read like a short spec.
- Keep blackbox coverage in `contract` and `integration`; keep branch-focused logic coverage in `unit`.
