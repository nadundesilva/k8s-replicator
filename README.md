# K8s Replicator 🚀

[![Main Branch Build](https://github.com/nadundesilva/k8s-replicator/actions/workflows/branch-build.yaml/badge.svg)](https://github.com/nadundesilva/k8s-replicator/actions/workflows/branch-build.yaml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[![Release](https://img.shields.io/github/release/nadundesilva/k8s-replicator.svg?style=flat-square)](https://github.com/nadundesilva/k8s-replicator/releases/latest)
[![Docker Image](https://img.shields.io/docker/image-size/nadunrds/k8s-replicator/latest?style=flat-square)](https://hub.docker.com/r/nadunrds/k8s-replicator)
[![Docker Pulls](https://img.shields.io/docker/pulls/nadunrds/k8s-replicator?style=flat-square)](https://hub.docker.com/r/nadunrds/k8s-replicator)

A Kubernetes operator for replicating resources across namespaces, designed with extensibility and performance in mind.

## 🔧 Supported Resources

**Currently Supported Resource Types:** See [API Documentation](API.md#supported-resources) for the complete list of supported resource types.

The system was designed with an extensible architecture that allows easy addition of new resource types as needed.

**Can't find the resource you're looking for?** 🚀 You can easily add support for any Kubernetes resource type! See our [Contributing Guide](CONTRIBUTING.md#extending-the-operator) for step-by-step instructions.

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

**Additional label-based capabilities**: The operator supports advanced namespace filtering, replica management, and more. See [API.md](API.md) for complete label and annotation reference.

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
- **[AI Agents Guide](AGENTS.md)** 🤖

> 💡 **Found an issue with our documentation?** We'd love your help! Please feel free to raise a pull request to improve it. Every contribution makes our docs better for everyone! 🤝✨

## Support 💬

- **Questions**: [Discussions](https://github.com/nadundesilva/k8s-replicator/discussions)
- **Bugs**: [Bug Report](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FBug&template=bug-report.md)
- **Features**: [Feature Request](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FFeature&template=feature-request.md)
