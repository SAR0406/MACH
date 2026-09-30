# MACH 0.1 MVP Contract

## Product thesis

MACH 0.1 proves one device can be treated as a VPS abstraction with stable identity, remote control, and public service exposure without requiring inbound networking setup.

## Scope (in)

1. Stable cryptographic device identity and claim flow
2. Outbound-only agent heartbeat and capability reporting
3. Public HTTP endpoint allocation model
4. Direct-first path with relay fallback path reporting
5. Command set: `init`, `status`, `expose`, `ssh`, `logs`, `stop`, `doctor`
6. Controller-side audit logs and token-authenticated device APIs

## Out of scope (0.1)

- Multi-node scheduler and capability-based placement
- Workload migration/failover
- Browser dashboard
- Full TLS automation implementation (endpoint model only for local MVP)
- QUIC transport hardening and federation

## Release acceptance criteria

- A fresh device can run `mach init`, receive a stable `device_id`, and persist identity in `~/.mach/v1/device.json`
- `mach-agent` reports heartbeat and capabilities to controller over outbound requests
- `mach expose <port>` returns canonical public URL and relay URL
- `mach status` displays device state and active network path (`DIRECT`/`RELAY`)
- Relay endpoint `/p/<id>` forwards to healthy online device exposures
- `mach doctor` reports identity, connectivity, path mode, DNS/TLS endpoint readiness, and guidance

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

### CLI command contract

- `mach init --name <name> --controller <url>`: creates keypair, claims device, stores config/token.
- `mach status`: prints state, path, heartbeat, capability snapshot, and exposures.
- `mach expose <port>`: validates local service, creates exposure, prints public + relay URLs.
- `mach ssh`: prints target SSH endpoint abstraction.
- `mach logs [-n N]`: prints recent controller audit entries for this device.
- `mach stop [exposure-id]`: stop one exposure or all exposures.
- `mach doctor`: prints diagnostic checks and recommended action.

### Exposure/session lifecycle states

- Device states: `CLAIMED` → `ONLINE` → `RECOVERING` / `OFFLINE`
- Path states: `UNKNOWN` → `DIRECT` or `RELAY`
- Exposure states: `PENDING` / `ACTIVE` / `STOPPED` (MVP actively uses `ACTIVE`)

## VPS Manifest v0

Defined in `proto/manifest.go`:
- `runtime.type` (`process`/`container` policy-owned)
- `network.http.port`
- `restart.policy`
- `health.command`

This keeps deployment contract minimal while exposure path stabilizes.

## Repository structure and engineering standards

- `agent/` — outbound agent binary
- `controller/` — API/control-plane binary
- `cli/` — user CLI binary
- `internal/api` — transport contracts + client
- `internal/controllerstate` — controller runtime state + persistence
- `internal/config` — schema paths and config storage
- `internal/identity` — key generation and signature verification
- `proto/` — workload manifest contracts

Standards:
- Logging: structured JSON audit entries for controller actions.
- Error model: plain-language API errors with proper HTTP status code classes.
- Config versioning: explicit `version` (`v1`) in device/controller state paths.
- State persistence: controller persists device and exposure state in `~/.mach-controller/v1/state/store.json`.

## Roadmap alignment

- 0.2: process/container supervision, resource limits, richer manifest
- 0.3: dashboard, domains, logs/metrics, browser terminal
- 0.4+: multi-node scheduling, migration, failover
