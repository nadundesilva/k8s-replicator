# K8s Replicator 🚀

[![Main Branch Build](https://github.com/nadundesilva/k8s-replicator/actions/workflows/branch-build.yaml/badge.svg)](https://github.com/nadundesilva/k8s-replicator/actions/workflows/branch-build.yaml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[![Release](https://img.shields.io/github/release/nadundesilva/k8s-replicator.svg?style=flat-square)](https://github.com/nadundesilva/k8s-replicator/releases/latest)
[![Docker Image](https://img.shields.io/docker/image-size/nadunrds/k8s-replicator/latest?style=flat-square)](https://hub.docker.com/r/nadunrds/k8s-replicator)
[![Docker Pulls](https://img.shields.io/docker/pulls/nadunrds/k8s-replicator?style=flat-square)](https://hub.docker.com/r/nadunrds/k8s-replicator)

A Kubernetes operator for replicating resources across namespaces, designed with extensibility and performance in mind.

**Supported Resources:** See [API Documentation](API.md#supported-resources) for currently supported resource types.

## Quick Start ⚡

### Install

```bash
curl -L https://raw.githubusercontent.com/nadundesilva/k8s-replicator/main/installers/install.sh | bash -s <VERSION>
```

### Mark Resource for Replication

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  labels:
    replicator.nadundesilva.github.io/object-type: replicated
```

### Uninstall

```bash
curl -L https://raw.githubusercontent.com/nadundesilva/k8s-replicator/main/installers/uninstall.sh | bash -s
```

## Documentation 📚

- **[Contributing](CONTRIBUTING.md)** 🤝
- **[API Reference](API.md)** 📖
- **[Troubleshooting](TROUBLESHOOTING.md)** 🔧
- **[Architecture](ARCHITECTURE.md)** 🏗️
- **[Examples](examples/)** 📚
- **[Benchmark Results](BENCHMARK.md)** ⚡
- **[Changelog](CHANGELOG.md)** 📝
- **[Security](SECURITY.md)** 🔒
- **[Adopters](ADOPTERS.md)** 🏢
- **[Code of Conduct](CODE_OF_CONDUCT.md)** 🤝

> 💡 **Found an issue with our documentation?** We'd love your help! Please feel free to raise a pull request to improve it. Every contribution makes our docs better for everyone! 🤝✨

## Support 💬

- **Questions**: [Discussions](https://github.com/nadundesilva/k8s-replicator/discussions)
- **Bugs**: [Bug Report](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FBug&template=bug-report.md)
- **Features**: [Feature Request](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FFeature&template=feature-request.md)
