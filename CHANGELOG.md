# Changelog

All notable changes to this project will be documented in this file.

- chore(deps): Go 1.26.5 and nox SDK v1.17.0 (#24)
- chore(security): nox remediation (deps + actions) (#23)
- ci: add nox-remediate caller (deps + action-pin remediation)
- ci: point the registry notice at where entries actually go (#22)
- ci: add nox self-scan and changed-files PR gate (#21)


## [0.7.0] - 2026-07-18

### Added

- **`drift` tool** — compares running workload state against declared IaC
  manifests, catching configuration applied out-of-band or diverging from what
  the repository says. Manifest review can pass while the live cluster differs.

  660+ lines that had previously existed only in nox's `plugins/` directory and
  had never been released: **v0.6.6 has no drift capability at all.**

### Changed

- `compareWorkload` takes its workload by pointer (600-byte struct, read-only
  inside), and finding loops index rather than copy 128 bytes per iteration.

### Note

No per-tool safety declarations (nox v1.9.1 `sdk.ToolSafety`), deliberately.
Both `scan` and `drift` talk to the cluster API, so neither is passive and
their requirements are identical — there is no narrowing to express, and
declaring one passive to make it run under a default policy would be a false
statement to the host.

## [0.1.0] - 2026-02-22

### Added

- Initial release with 8 Kubernetes runtime security rules (KRUNT-001 through KRUNT-008)
- Live cluster scanning via Kubernetes API (in-cluster and kubeconfig support)
- Container-level checks: root execution, privileged mode, resource limits, unpinned images, dangerous capabilities
- Pod-level checks: host namespace sharing, network policy enforcement, service account token auto-mount
- Compliance mappings: CIS, NIST-800-53, PCI-DSS, OWASP Top 10
