# Agent context index — stolostron/azure-service-operator

Entry point for AI coding agents (and new humans) working in this repository.

This file is a **table of contents**, not a manual. Read the linked document
for detail before making changes. For fork-specific context (branch model,
upstream sync, build pipelines) start with
[`docs/DOWNSTREAM.md`](docs/DOWNSTREAM.md).

## What this repository is

`stolostron/azure-service-operator` is a downstream fork of
[`Azure/azure-service-operator`](https://github.com/Azure/azure-service-operator)
(ASO). It packages the ASO controller for **Azure Red Hat OpenShift Hosted
Control Planes (ARO-HCP)** and ships it as part of **multicluster engine
(MCE) / backplane** via [Konflux](https://konflux-ci.dev/) builds.

## Where to start

| Document | What it covers |
|----------|----------------|
| [README.md](README.md) | Project overview. |
| [v2/README.md](v2/README.md) | ASO v2 installation, usage, and supported resources. |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Upstream contribution workflow. |
| [docs/DOWNSTREAM.md](docs/DOWNSTREAM.md) | **Fork-specific context** — branch model, upstream sync, Konflux pipelines, what lives in `stolostron/`. Read this before opening a PR. |
| [docs/branch-sync-policy.md](docs/branch-sync-policy.md) | Authoritative branch-to-source mapping and protection rules. |
| [SECURITY.md](SECURITY.md) | Vulnerability reporting. |
| [SUPPORT.md](SUPPORT.md) | Support channels. |

## Key directories

| Path | Purpose |
|------|---------|
| `v2/` | All Go source (controller, CRDs, API types, tests). |
| `v2/api/` | Generated CRD API types. |
| `v2/pkg/` | Controller implementation packages. |
| `stolostron/` | Downstream-only build artifacts (`Dockerfile.stolostron`, `Makefile`). |
| `.tekton/` | Konflux `PipelineRun` definitions for MCE releases. |
| `.github/workflows/` | CI workflows including upstream sync and branch fast-forwards. |
| `docs/` | Human-readable documentation and policies. |
