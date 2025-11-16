# Documentation Migration Checklist — KSail to KSail-Go

| Source Path | Target Path | Category | Owner | Status | Command Updates? | Notes |
|-------------|-------------|----------|-------|--------|-------------------|-------|
| `projects/ksail/docs/overview/index.md` | `docs/overview/index.md` | Overview | Docs | Migrated | Yes | Copy rewritten for KSail-Go positioning with updated navigation |
| `projects/ksail/docs/overview/project-structure.md` | `docs/overview/project-structure.md` | Overview | Docs | Migrated | Yes | Repository references and cross-links updated |
| `projects/ksail/docs/overview/support-matrix.md` | `docs/overview/support-matrix.md` | Overview | Docs | Migrated | No | Distribution matrix refreshed and linted |
| `projects/ksail/docs/overview/core-concepts/index.md` | `docs/overview/core-concepts/index.md` | Core Concepts | Docs | Migrated | Yes | Navigation between subtopics reworked |
| `projects/ksail/docs/overview/core-concepts/cnis.md` | `docs/overview/core-concepts/cnis.md` | Core Concepts | Docs | Migrated | Yes | KSail-Go commands replace legacy syntax |
| `projects/ksail/docs/overview/core-concepts/container-engines.md` | `docs/overview/core-concepts/container-engines.md` | Core Concepts | Docs | Migrated | Yes | Kind/K3d instructions aligned with Go CLI |
| `projects/ksail/docs/overview/core-concepts/csis.md` | `docs/overview/core-concepts/csis.md` | Core Concepts | Docs | Migrated | Yes | Storage examples validated |
| `projects/ksail/docs/overview/core-concepts/deployment-tools.md` | `docs/overview/core-concepts/deployment-tools.md` | Core Concepts | Docs | Migrated | Yes | Flux/kubectl coverage updated |
| `projects/ksail/docs/overview/core-concepts/distributions.md` | `docs/overview/core-concepts/distributions.md` | Core Concepts | Docs | Migrated | Yes | Supported distributions table refreshed |
| `projects/ksail/docs/overview/core-concepts/editor.md.md` | `docs/overview/core-concepts/editor.md` | Core Concepts | Docs | Migrated | No | Duplicate extension removed during migration |
| `projects/ksail/docs/overview/core-concepts/gateway-controllers.md` | `docs/overview/core-concepts/gateway-controllers.md` | Core Concepts | Docs | Migrated | Yes | Feature availability updated |
| `projects/ksail/docs/overview/core-concepts/index.md` | `docs/overview/core-concepts/index.md` | Core Concepts | Docs | Migrated | Yes | Table of contents consolidated |
| `projects/ksail/docs/overview/core-concepts/ingress-controllers.md` | `docs/overview/core-concepts/ingress-controllers.md` | Core Concepts | Docs | Migrated | Yes | Terminology aligned with Go CLI |
| `projects/ksail/docs/overview/core-concepts/local-registry.md` | `docs/overview/core-concepts/local-registry.md` | Core Concepts | Docs | Migrated | Yes | Registry configuration steps updated |
| `projects/ksail/docs/overview/core-concepts/metrics-server.md` | `docs/overview/core-concepts/metrics-server.md` | Core Concepts | Docs | Migrated | Yes | Metrics installation flow verified |
| `projects/ksail/docs/overview/core-concepts/mirror-registries.md` | `docs/overview/core-concepts/mirror-registries.md` | Core Concepts | Docs | Migrated | Yes | Mirrored registry procedure confirmed |
| `projects/ksail/docs/overview/core-concepts/secret-manager.md` | `docs/overview/core-concepts/secret-manager.md` | Core Concepts | Docs | Migrated | Yes | SOPS integration updated for Go CLI |
| `projects/ksail/docs/configuration/index.md` | `docs/configuration/index.md` | Configuration | Docs | Migrated | Yes | Navigation and precedence sections refreshed |
| `projects/ksail/docs/configuration/cli-options.md` | `docs/configuration/cli-options.md` | Configuration | Docs | Migrated | Yes | Command flags rewritten and linted |
| `projects/ksail/docs/configuration/declarative-config.md` | `docs/configuration/declarative-config.md` | Configuration | Docs | Migrated | Yes | YAML examples aligned with ksail-go |
| `projects/ksail/docs/use-cases/index.md` | `docs/use-cases/index.md` | Use Cases | Docs | Migrated | Yes | Overview language refreshed and cross-linked |
| `projects/ksail/docs/use-cases/local-development.md` | `docs/use-cases/local-development.md` | Use Cases | Docs | Migrated | Yes | Quick start commands updated for Go CLI |
| `projects/ksail/docs/use-cases/learning-kubernetes.md` | `docs/use-cases/learning-kubernetes.md` | Use Cases | Docs | Migrated | Yes | Tutorial references Go CLI workflow |
| `projects/ksail/docs/use-cases/e2e-testing-in-cicd.md` | `docs/use-cases/e2e-testing-in-cicd.md` | Use Cases | Docs | Migrated | Yes | CI workflow steps validated |
| `projects/ksail/docs/images/architecture.drawio.png` | `docs/images/architecture.drawio.png` | Asset | Docs | Copied | No | Diagram copied; references fixed |
| `projects/ksail/docs/images/enable-docker-socket-in-docker-desktop.png` | `docs/images/enable-docker-socket-in-docker-desktop.png` | Asset | Docs | Copied | No | Screenshot copied from legacy docs |
| `projects/ksail/docs/images/github-mark-white.png` | `docs/images/github-mark-white.png` | Asset | Docs | Copied | No | Asset copied |
| `projects/ksail/docs/images/gitops-structure.drawio.png` | `docs/images/gitops-structure.drawio.png` | Asset | Docs | Copied | No | Diagram copied |
| `projects/ksail/docs/images/ksail-cli-dark.png` | `docs/images/ksail-cli-dark.png` | Asset | Docs | Copied | Maybe | Legacy screenshot retained; revisit when CLI output changes |
| `projects/ksail/docs/images/ksail-cli-light.png` | `docs/images/ksail-cli-light.png` | Asset | Docs | Copied | Maybe | Legacy screenshot retained; revisit when CLI output changes |
| `projects/ksail/docs/images/ksail-logo.png` | `docs/images/ksail-logo.png` | Asset | Docs | Copied | No | Logo copied |
| `projects/ksail/docs/faq.md` | – | Deferred | | Deferred | – | Legacy content references .NET; evaluate separately |
| `projects/ksail/docs/roadmap.md` | – | Deferred | | Deferred | – | Out of scope for migration |
| `projects/ksail/docs/404.md` | – | Deferred | | Deferred | – | Jekyll-specific page excluded |
