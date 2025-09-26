# Contributing to K8s Replicator 🚀

Thank you for your interest in contributing to K8s Replicator! 🎉 This guide will help you get started with development setup, workflow, and the submission process.

## Quick Start 🏁

### Prerequisites

Before you begin, ensure you have the following tools installed:

- **Go** 🐹 (latest stable version)
- **Docker** 🐳 or **Podman** for container builds
- **kubectl** ☸️ for Kubernetes cluster interaction
- **Operator SDK** ⚙️ (latest stable version)

### Setup

```bash
# Fork and clone the repository
git clone https://github.com/your-username/k8s-replicator.git
cd k8s-replicator
git remote add upstream https://github.com/nadundesilva/k8s-replicator.git

# Install all required dependencies
make controller-gen kustomize envtest golangci-lint operator-sdk

# Generate necessary code artifacts
make manifests generate

# Run the controller locally for development
make run
```

## Development Workflow 🔄

### 1. Create Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

- Write clean, well-documented code 📝
- Add tests for new functionality 🧪
- Update documentation as needed 📚
- **Update release notes** 📝 in `.github/RELEASE_NOTE` file with your changes

### 3. Code Quality

```bash
make fmt lint vet
```

### 4. Testing

```bash
make test              # All tests
make test.unit         # Unit tests only
make test.e2e          # E2E tests only
make test.benchmark    # Benchmark tests
```

## Building & Deployment 🏗️

### Build

```bash
make build              # Binary
make docker-build       # Docker image
make docker-buildx      # Multi-platform
```

### Deploy

```bash
make install deploy     # Local deployment
make undeploy          # Cleanup
```

### Bundle (OLM)

```bash
make bundle            # Generate bundle
make bundle-build       # Build bundle image
make bundle-push        # Push bundle image
```

## Extending the Operator 🔧

The operator uses a `Replicator` interface for extensibility:

```go
type Replicator interface {
    GetKind() string
    AddToScheme(scheme *runtime.Scheme) error
    EmptyObject() client.Object
    EmptyObjectList() client.ObjectList
    ObjectListToArray(client.ObjectList) []client.Object
    Replicate(sourceObject client.Object, targetObject client.Object)
}
```

### Adding New Resource Types

The extensible architecture makes it easy to add support for any Kubernetes resource:

1. **Create Implementation**: Add new replicator in `controllers/replication/`
2. **Implement Interface**: Implement the `Replicator` interface for your resource type
3. **Register**: Add to `NewReplicators()` function in `controllers/replication/replicator.go`
4. **Test & Document**: Add tests and update documentation
5. **Update API Documentation**: Add the new resource type to the [Supported Resources](API.md#supported-resources) section in `API.md`

**Benefits of Extensible Design:**

- **No Core Changes**: Adding resources doesn't require modifying core logic
- **Independent Development**: Each resource type can be developed separately
- **Easy Testing**: Each replicator can be tested in isolation
- **Future-Proof**: Works with any current or future Kubernetes resource

## Submitting Changes 📤

### 1. Commit Changes

```bash
git add .
git commit -m "feat: add support for YourResource type"
```

**Commit Types:**

- `feat:` new features
- `fix:` bugfixes
- `docs:` documentation
- `test:` tests
- `refactor:` refactoring
- `perf:` performance

### 2. Create Pull Request

```bash
git push origin feature/your-feature-name
```

**PR Guidelines:**

- One feature per PR 🎯
- Include tests 🧪
- Update documentation 📚
- Update release notes 📝
- Follow code style 🎨

## Getting Help 💬

- **Discussions**: [GitHub Discussions](https://github.com/nadundesilva/k8s-replicator/discussions)
- **Bugs**: [Bug Report](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FBug&template=bug-report.md)
- **Features**: [Feature Request](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FFeature&template=feature-request.md)

---

Thank you for contributing! 🎉✨
