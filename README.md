# MACH

MACH (Machine-as-a-Cloud Host) turns a single owned device into a VPS-like node.

## MACH 0.1 MVP

MACH 0.1 enforces these non-negotiables:
- One stable device identity
- One controllable agent
- One public HTTP exposure path
- Direct-or-relay connectivity reporting
- Core CLI lifecycle (`init`, `status`, `expose`, `ssh`, `logs`, `stop`) plus `doctor`

## Repository Layout

- `/agent` — agent binary
- `/controller` — control plane binary
- `/cli` — CLI binary
- `/proto` — API and manifest contracts
- `/docs` — MVP docs

## Quick start

```bash
make build
./bin/mach-controller
./bin/mach init --name my-machine
./bin/mach-agent
./bin/mach expose 3000
./bin/mach status
./bin/mach doctor
```

## Notes

- `https://<id>.mach.dev` is allocated as the canonical endpoint model.
- Local development relay URL is emitted as `http://127.0.0.1:8080/p/<id>`.
- Controller state is persisted at `~/.mach-controller/v1/state/store.json`.
- Product contract and MVP acceptance criteria live in `/home/runner/work/MACH/MACH/docs/mvp.md`.
