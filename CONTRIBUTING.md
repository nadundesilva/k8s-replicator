# Contributing to K8s Replicator 🚀

Thank you for your interest in contributing to K8s Replicator! 🎉 This guide will help you get started with development setup, workflow, and the submission process.

## Quick Start 🏁

### Prerequisites

Before you start, you'll need these tools:

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
make test.e2e          # End-to-end tests only
make test.benchmark    # Benchmark tests
```

#### Running Tests for Specific Resources

To run end-to-end tests for a specific resource type only, use the `TEST_RESOURCES_FILTER_REGEX` environment variable:

```bash
TEST_RESOURCES_FILTER_REGEX="<ResourceName>" make test.e2e
```

**Note**: The `<ResourceName>` should match the name returned by the test data generation function (e.g., `ServiceAccount`, `Secret`, `ConfigMap`, `NetworkPolicy`).

This is particularly useful when:

- **Debugging specific resource issues** 🐛
- **Testing new resource types** during development
- **Faster iteration** when working on a single resource type
- **CI/CD optimization** for targeted testing

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

The operator uses a `Replicator` interface for extensibility. See the complete interface definition and documentation in [`controllers/replication/replicator.go`](controllers/replication/replicator.go) or the [API Documentation](API.md#replicator-interface-).

### Adding New Resource Types

The extensible architecture makes it easy to add support for any Kubernetes resource:

1. **Create Implementation**: Add new replicator in `controllers/replication/`
2. **Implement Interface**: Implement the `Replicator` interface for your resource type
3. **Register**: Add to `NewReplicators()` function in `controllers/replication/replicator.go`
4. **Update RBAC**: Add kubebuilder RBAC comments in the new replicator file in `controllers/replication/`
5. **Update GitHub Actions**: Add new resource type to end-to-end test matrix in `.github/workflows/build.yaml`
6. **Test & Document**: Add tests and update documentation
7. **Update API Documentation**: Add the new resource type to the [Supported Resources](API.md#supported-resources) section in `API.md`
8. **Update Release Notes**: Add your new feature to `.github/RELEASE_NOTE` for the next release

**GitHub Actions End-to-End Test Matrix Update:**

When adding new resource types, you must update the end-to-end test matrix in `.github/workflows/build.yaml`:

```yaml
# In the run-e2e-tests job (around line 172-176)
strategy:
  matrix:
    resource:
      - Secret
      - ConfigMap
      - NetworkPolicy
      - ServiceAccount # Add your new resource type here
```

This ensures the new resource type is tested in the CI/CD pipeline.

**Benefits of Extensible Design:**

- **No Core Changes**: Adding resources doesn't require modifying core logic
- **Independent Development**: Each resource type can be developed separately
- **Easy Testing**: Each replicator can be tested in isolation (pure data transformation)
- **Clean Architecture**: Replication logic is separate from API operations
- **Future-Proof**: Works with any current or future Kubernetes resource

## Submitting Changes 📤

### 1. Commit Changes

```bash
git add .
git commit -m "feat: add support for YourResource type"
```

**Commit Format:**

This project uses **semantic release commit format**.

**Commit Message Format:**

```text
<type>(optional scope): <description>

[optional body]

[optional footer(s)]
```

**Types:**

- `feat:` A new feature
- `fix:` A bugfix
- `docs:` Documentation only changes
- `style:` Changes that do not affect the meaning of the code
- `refactor:` A code change that neither fixes a bug nor adds a feature
- `perf:` A code change that improves performance
- `test:` Adding missing tests or correcting existing tests
- `chore:` Changes to the build process or auxiliary tools
- `ci:` Changes to our CI configuration files and scripts
- `build:` Changes that affect the build tool or external dependencies
- `revert:` Reverts a previous commit

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
- **Update [AGENTS.md](AGENTS.md) if needed** 🤖 - Review and update the AI agents navigation guide when:
  - Adding new documentation files
  - Changing file locations or architecture
  - Modifying development workflows

## AI Agent Usage Policy 🤖

**AI agents are welcome and acceptable** for contributing to this project! However, we encourage responsible usage:

### ✅ **Encouraged Practices:**

- **Read the documentation first**: Understand the project through [README.md](README.md), [ARCHITECTURE.md](ARCHITECTURE.md), and [API.md](API.md)
- **Use [AGENTS.md](AGENTS.md)**: Follow the AI agent navigation guide for efficient project understanding
- **Verify outputs**: Review and test AI-generated code before submitting
- **Follow project patterns**: Make sure AI contributions match our existing code style and architecture

### ⚠️ **Important Guidelines:**

- **Understanding is required**: Don't submit contributions without understanding the project's purpose and design
- **Quality over speed**: Take time to make sure contributions are correct and well-tested
- **Human oversight**: Always review AI-generated code for correctness and adherence to project standards
- **Documentation matters**: Update relevant documentation when making changes

**Remember**: AI agents are tools to enhance productivity, but understanding the project and maintaining quality standards remains essential! 🎯

## Getting Help 💬

For support and help options, see the main [Support section](README.md#support-) in the project documentation.

---

Thank you for contributing! 🎉✨
