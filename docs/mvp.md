# MACH 0.1 MVP Contract

## Scope (non-negotiable)

1. One stable device identity
2. One controllable outbound agent
3. One public HTTP exposure path
4. Direct-first with relay fallback reporting
5. Core CLI commands: `init`, `status`, `expose`, `ssh`, `logs`, `stop`
6. Reliability command: `doctor`

## Identity and Security

- Device keypair generated locally and stored in `~/.mach/v1/device.json` (mode `0600`).
- Claim flow uses a signed challenge (`/v1/register/challenge` + `/v1/register/complete`).
- All device control APIs require bearer session tokens.
- Agent requires outbound-only connectivity; no inbound listening ports are required.

## Networking model

- Controller evaluates direct reachability to the advertised host and sets path as `DIRECT` when reachable.
- If direct checks fail, path is `RELAY`.
- Canonical endpoint model allocates `https://<id>.mach.dev`.
- Dev relay path is served at `http://<controller>/p/<id>`.

## APIs

- Register challenge
- Register complete
- Heartbeat
- Expose service
- Fetch status
- Stop exposure
- Fetch logs

## VPS Manifest v0

Defined in `proto/manifest.go`:
- `runtime.type` (`process`/`container` policy-owned)
- `network.http.port`
- `restart.policy`
- `health.command`

This keeps deployment contract minimal while exposure path stabilizes.

## Roadmap alignment

- 0.2: process/container supervision, resource limits, richer manifest
- 0.3: dashboard, domains, logs/metrics, browser terminal
- 0.4+: multi-node scheduling, migration, failover
