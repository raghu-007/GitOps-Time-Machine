# GitOps-Time-Machine ⏰🔀📦

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://img.shields.io/badge/CI-passing-brightgreen)]()

> **A system that continuously versions the actual state of live infrastructure into a Git repository, enabling time-travel debugging and drift analysis.**

---

## 🚀 What is GitOps-Time-Machine?

GitOps-Time-Machine captures **point-in-time snapshots** of your Kubernetes infrastructure and stores them as version-controlled YAML files in a Git repository. This enables you to:

- **⏪ Time-Travel Debug** — See exactly what your infrastructure looked like at any point in time
- **🔍 Drift Detection** — Compare live state against snapshots to detect unauthorized or accidental changes
- **📊 Diff Analysis** — Compare any two snapshots to understand what changed between them
- **📜 Audit Trail** — Maintain a complete, Git-backed history of infrastructure changes
- **⏰ Continuous Monitoring** — Schedule automatic snapshots on a cron schedule

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────┐
│                    CLI (Cobra)                   │
│  snapshot │ diff │ drift │ history │ watch       │
├───────────┬───────────┬─────────────────────────┤
│ Collector │Snapshotter│ Git Versioner            │
│ (K8s API) │(YAML/disk)│ (go-git)                │
├───────────┴───────────┴─────────────────────────┤
│         Drift Analyzer  │  Time-Travel Query     │
├─────────────────────────┴───────────────────────┤
│           Scheduler (cron)  │  Config (Viper)    │
└─────────────────────────────────────────────────┘
```

| Component | Description |
|-----------|-------------|
| **Collector** | Connects to Kubernetes via `client-go` dynamic client, discovers and fetches configured resource types |
| **Snapshotter** | Serializes resources to organized YAML files (`namespace/kind/name.yaml`), handles field stripping |
| **Git Versioner** | Manages the snapshot Git repo — init, commit with metadata, history log, time-based checkout |
| **Drift Analyzer** | Deep-compares two snapshots with field-level diff detection |
| **Time-Travel Engine** | Resolves timestamps to Git commits, enables historical state queries |
| **Scheduler** | Cron-based scheduling for continuous snapshot capture |

---

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/raghu-007/GitOps-Time-Machine.git
cd GitOps-Time-Machine

# Build the binary
go build -o bin/gitops-time-machine .

# Or use Make
make build
```

### Docker

```bash
docker build -t gitops-time-machine .
docker run --rm -v ~/.kube:/root/.kube gitops-time-machine snapshot
```

---

## ⚡ Quick Start

### 1. Configure

```bash
# Copy the example config
cp config.example.yaml config.yaml

# Edit to match your environment
# At minimum, ensure kubeconfig path is correct
```

### 2. Take Your First Snapshot

```bash
./bin/gitops-time-machine snapshot
```

Output:
```
  ╔══════════════════════════════════════════════╗
  ║       GitOps-Time-Machine  ⏰ → 🔀 → 📦      ║
  ║   Infrastructure Time-Travel & Drift Detect  ║
  ╚══════════════════════════════════════════════╝

📸 Snapshot Captured
─────────────────────────────────────────────
  ⏰  Time:       2024-01-15 10:30:00 UTC
  🏗️  Cluster:    production
  📦  Resources:  142
  🗂️  Namespaces: 8
  🔗  Commit:     a1b2c3d4
```

### 3. Detect Drift

```bash
./bin/gitops-time-machine drift
```

### 4. View History

```bash
./bin/gitops-time-machine history --limit 10
```

### 5. Compare Snapshots

```bash
# Compare by timestamps
./bin/gitops-time-machine diff \
  --from "2024-01-01T00:00:00Z" \
  --to "2024-01-15T00:00:00Z"

# Compare with a specific commit
./bin/gitops-time-machine diff --commit a1b2c3d4
```

### 6. Continuous Monitoring

```bash
# Watch with default schedule (every 5 minutes)
./bin/gitops-time-machine watch

# Custom schedule (every hour)
./bin/gitops-time-machine watch --schedule "0 * * * *"
```

---

## 🛠️ CLI Reference

| Command | Description |
|---------|-------------|
| `snapshot` | Capture a one-time infrastructure snapshot |
| `diff` | Compare two snapshots by time or commit |
| `drift` | Detect drift between live state and last snapshot |
| `history` | List all committed snapshots |
| `watch` | Start continuous scheduled snapshotting |
| `version` | Print version information |

### Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Path to config file (default: `./config.yaml`) |
| `--kubeconfig` | Path to kubeconfig file |
| `-v, --verbose` | Enable debug logging |

---

## ⚙️ Configuration

GitOps-Time-Machine supports configuration via YAML file, environment variables (`GTM_` prefix), and CLI flags (highest priority).

See [`config.example.yaml`](config.example.yaml) for all available options.

### Key Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `snapshot.output_dir` | `./infra-snapshots` | Where to store snapshots |
| `snapshot.resource_types` | Core K8s resources | Which resource types to capture |
| `snapshot.exclude_namespaces` | `kube-system`, `kube-public`, `kube-node-lease` | Namespaces to skip |
| `git.branch` | `main` | Branch for the snapshot repo |
| `watch.schedule` | `*/5 * * * *` | Cron schedule for continuous mode |

---

## 📁 Snapshot Directory Structure

```
infra-snapshots/
├── .git/
├── _metadata.yaml
├── _cluster/
│   ├── clusterrole/
│   │   ├── admin.yaml
│   │   └── viewer.yaml
│   └── clusterrolebinding/
│       └── admin-binding.yaml
├── default/
│   ├── deployment/
│   │   ├── nginx.yaml
│   │   └── api-server.yaml
│   ├── service/
│   │   └── nginx-svc.yaml
│   └── configmap/
│       └── app-config.yaml
└── monitoring/
    ├── deployment/
    │   └── prometheus.yaml
    └── service/
        └── prometheus-svc.yaml
```

---

## 🗺️ Roadmap

- [ ] Webhook/alerting integration (Slack, PagerDuty)
- [ ] Web UI dashboard for visual time-travel
- [ ] Support for non-Kubernetes infrastructure (AWS, GCP, Azure)
- [ ] Terraform state snapshotting
- [ ] RBAC-aware secret masking
- [ ] Prometheus metrics endpoint
- [ ] Helm chart for in-cluster deployment

---

## 🤝 Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## 📄 License

This project is licensed under the **Apache License 2.0** — see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Viper](https://github.com/spf13/viper) — Configuration management
- [go-git](https://github.com/go-git/go-git) — Pure Go Git implementation
- [client-go](https://github.com/kubernetes/client-go) — Kubernetes client library
- [Logrus](https://github.com/sirupsen/logrus) — Structured logging
