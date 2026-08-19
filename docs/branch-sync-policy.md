# Branch synchronization & immutability policy

This document is the **source of truth** for how branches in
`stolostron/azure-service-operator` are sourced, synchronized, and protected.
There is no team wiki, so the branch-to-source mapping lives here in the repo.

Tracking: [ARO-29174](https://redhat.atlassian.net/browse/ARO-29174) (follows the
branch reorganization in ARO-27857).

## Branch-to-source mapping

| Branch | Source / basis | Sync mechanism | Immutability |
|---|---|---|---|
| `main` | upstream `Azure/main` + ARO-HCP (2026-06-30-preview) customizations | Periodic **merge** from `Azure/main` via [`sync-upstream-main.yaml`](../.github/workflows/sync-upstream-main.yaml) (opens a PR — not a fast-forward, because `main` carries ARO-HCP customizations) | Open for PRs (default branch) |
| `backplane-5.1` | `stolostron/main` | **FFWD only** from `main` via [`ffwd-branch.yaml`](../.github/workflows/ffwd-branch.yaml) | No direct PRs (blocked by [`protect-backplane-branches.yaml`](../.github/workflows/protect-backplane-branches.yaml) + ruleset) |
| `backplane-5.0` | `stolostron/release-2.18` | **FFWD only, one-directional** from `release-2.18` via [`ffwd-release-2.18.yaml`](../.github/workflows/ffwd-release-2.18.yaml) | No direct PRs (blocked by `protect-backplane-branches.yaml` + ruleset) |
| `release-2.19` | upstream v2.19 + ARO-HCP | Manual, security updates only | Immutable except security updates |
| `release-2.18` | upstream v2.18 + ARO-HCP | Manual, security updates only (also feeds `backplane-5.0`) | Immutable except security updates |
| `release-2.13` | upstream v2.13 + ARO-HCP (**former `main`**) | Manual, security updates only | Immutable except security findings |
| `backplane-2.17` | — | Manual, security findings only | Immutable except security findings |
| `backplane-2.11` | — | Manual, security findings only | Immutable except security findings |

## Synchronization workflows

- **`main` ← `Azure/main`** — [`sync-upstream-main.yaml`](../.github/workflows/sync-upstream-main.yaml)
  runs weekly (and on demand). Because `main` carries ARO-HCP customizations, it
  performs a **merge** on a `sync/upstream-main` branch and opens a PR for review.
  Conflicts fail the job so they are resolved manually. Nothing is force-pushed to `main`.
- **`main` → `backplane-5.1`** — [`ffwd-branch.yaml`](../.github/workflows/ffwd-branch.yaml)
  fast-forwards every push to `main` into `backplane-5.1`. One-directional.
- **`release-2.18` → `backplane-5.0`** — [`ffwd-release-2.18.yaml`](../.github/workflows/ffwd-release-2.18.yaml)
  fast-forwards every push to `release-2.18` into `backplane-5.0`. One-directional.

Fast-forward is deliberate: the `backplane-*` targets must never diverge from
their source. Direct PRs to `backplane-5*` are rejected by
`protect-backplane-branches.yaml`; changes flow only through the source branch.

## Immutability / branch protection (rulesets)

Enforced via GitHub repository **rulesets** (applied through the admin API). The
required posture:

| Branch(es) | Ruleset enforcement |
|---|---|
| `backplane-5.0`, `backplane-5.1` | Non-fast-forward blocked, deletion blocked; updates only via the FFWD bots. Direct PRs rejected by `protect-backplane-branches.yaml`. |
| `release-2.13`, `backplane-2.17`, `backplane-2.11` | Immutable: no direct pushes, no deletion, no force-push. Changes only for security findings, via reviewed PR. |
| `release-2.18`, `release-2.19` | Protected: no force-push, no deletion. Security updates only, via reviewed PR. |
| `main` | Default protections (PR review + status checks); receives upstream via the sync PR. |

> The mapping table above is authoritative. If a ruleset and this table ever
> disagree, update whichever is wrong so they match.

## Audit — customizations preserved after `main` → `release-2.13`

The previous `main` (now `release-2.13`) accumulated repo/branch-level
configuration. When `main` was re-seeded from `Azure/main`, all of the following
were verified to be carried onto the new `main` so nothing was silently lost:

- **GitHub Actions workflows** — full `.github/workflows/` set (CI, PR validation,
  release, security scanning, scorecards, CodeQL, FFWD sync, etc.).
- **CodeRabbit** — `.coderabbit.yaml`.
- **Renovate** — `renovate.json` + `.github/workflows/renovate.yaml`.
- **Dependabot** — `.github/dependabot.yml`.
- **CODEOWNERS** — `.github/CODEOWNERS`.
- **Copilot / PR templates** — `.github/copilot-instructions.md`,
  `.github/pull_request_template.md`, `.github/ISSUE_TEMPLATE/`.
- **Konflux / Tekton** — `.tekton/` pipelines.
- **stolostron downstream carry** — `v2/stolostron/` CRDs bundle, Dockerfiles,
  governance and samples.

Branch rulesets and protection are repository-level settings (not files), so they
are reapplied via the admin API per the table above rather than carried by the
branch move.
