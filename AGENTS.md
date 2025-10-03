# AI Agent Navigation Guide 🤖

> **Quick orientation for AI assistants working with K8s Replicator**

## Project Type & Context 📋

**K8s Replicator** is a Kubernetes operator that replicates resources across namespaces using labels/annotations. See [README.md](README.md) for project overview and [ARCHITECTURE.md](ARCHITECTURE.md) for technical details.

**Key Concept**: Mark resources with `replicator.nadundesilva.github.io/object-type: replicated` to replicate them across namespaces.

## Quick Start for AI Agents 🚀

**First Time Here?** Follow this sequence:

1. **Read**: [README.md](README.md) for project overview (5 min)
2. **Understand**: [ARCHITECTURE.md](ARCHITECTURE.md) for system design (10 min)
3. **Explore**: `controllers/replication/replicator.go` for core interface (5 min)
4. **Reference**: [API.md](API.md) for supported resources and labels

**Need to make changes?** See [CONTRIBUTING.md](CONTRIBUTING.md#development-workflow-) for complete workflow.

## Essential Documentation Map 🗺️

### Start Here First

- **[README.md](README.md)** - Project overview, quick start, installation
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Technical architecture, controllers, data flow diagrams

### For Code Analysis & Development

- **[API.md](API.md)** - Complete API reference, supported resources (definitive list), labels/annotations
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development setup, adding new resource types, testing
- **Core implementation**: `controllers/` directory
  - `replication/replicator.go` - Main interface with full Go docstrings
  - `replication_controller.go` - Resource replication logic
  - `namespace_controller.go` - Namespace lifecycle management

### For Troubleshooting & Debugging

- **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Common issues, debugging steps
- **Logs location**: `k8s-replicator-system` namespace
- **Key files**: `config/rbac/` for permissions, `config/samples/` for examples

### For Understanding Usage

- **[examples/](examples/)** - Practical examples with setup instructions
- **[examples/cert-manager/](examples/cert-manager/)** - Complete integration example

## Key Architecture Points 🏗️

### Independent Controllers

- **Replication Controller**: Watches source resources → replicates to namespaces
- **Namespace Controller**: Watches namespaces → replicates existing source resources to new namespaces
- **No direct communication** between controllers
- **Critical insight**: Namespace Controller replicates ALL existing source resources to new namespaces automatically

**For detailed architecture**: See [ARCHITECTURE.md](ARCHITECTURE.md) for comprehensive system design and data flow diagrams.

### Core Interface Pattern

- **Complete interface definition**: `internal/controller/replication/replicator.go` (with full Go docstrings)
- **Implementation examples**: `internal/controller/replication/*.go` files
- **Critical insight**: The `Replicate()` method performs ONLY in-memory data copying - no API calls
- **Key pattern**: Each resource type has its own replicator file (e.g., `secret.go`, `serviceaccount.go`)

**For interface details**: See `controllers/replication/replicator.go` for complete interface documentation and [API.md](API.md) for API reference.

### Labels/Annotations System

- **Complete reference**: [API.md](API.md) - all labels, annotations, and values

## AI Agent Guidelines 🎯

### When Analyzing Code

1. **Start with**: [ARCHITECTURE.md](ARCHITECTURE.md) for system understanding
2. **Interface details**: `controllers/replication/replicator.go` (has comprehensive docstrings)
3. **Controller logic**: `*_controller.go` files in `controllers/`
4. **RBAC requirements**: `config/rbac/role.yaml`

### When Making Changes

1. **Adding new resource types**: See [CONTRIBUTING.md](CONTRIBUTING.md#adding-new-resource-types) for step-by-step guide
2. **Development workflow**: See [CONTRIBUTING.md](CONTRIBUTING.md#development-workflow-) for complete process
3. **Testing requirements**: See [CONTRIBUTING.md](CONTRIBUTING.md#testing) for test commands and structure
4. **Testing specific resources**: See [CONTRIBUTING.md](CONTRIBUTING.md#running-tests-for-specific-resources) for targeted testing with `TEST_RESOURCES_FILTER_REGEX`
5. **Critical steps**: Always run `make bundle` after adding kubebuilder RBAC comments
6. **GitHub Actions**: Update end-to-end test matrix in `.github/workflows/build.yaml` for new resource types

### When Debugging Issues

1. **Check**: [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common problems
2. **Verify**: Labels/annotations using [API.md](API.md) reference
3. **Logs**: Controller logs in `k8s-replicator-system` namespace
4. **Examples**: [examples/](examples/) for working configurations

### When Explaining to Users

1. **Quick start**: Direct to [README.md](README.md)
2. **Setup help**: Point to [examples/](examples/) directory
3. **Technical questions**: Reference [ARCHITECTURE.md](ARCHITECTURE.md)
4. **API questions**: Always link to [API.md](API.md)

## Common AI Agent Scenarios 🎯

### "I need to add support for a new resource type"

1. **Read**: [CONTRIBUTING.md](CONTRIBUTING.md#adding-new-resource-types) for complete guide
2. **Create**: New replicator in `controllers/replication/`
3. **Register**: Add to `NewReplicators()` in `controllers/replication/replicator.go`
4. **RBAC**: Add kubebuilder comments in `controllers/replication_controller.go`
5. **Test**: Create test data in `test/utils/testdata/`
6. **Bundle**: Run `make bundle` to generate RBAC
7. **CI/CD**: Update `.github/workflows/build.yaml` end-to-end matrix
8. **Docs**: Update [API.md](API.md) supported resources list

### "I need to understand how replication works"

1. **Start**: [ARCHITECTURE.md](ARCHITECTURE.md) for system overview
2. **Interface**: `controllers/replication/replicator.go` for contract
3. **Implementation**: `controllers/replication/*.go` for examples
4. **Controllers**: `controllers/replication_controller.go` and `controllers/namespace_controller.go`

### "I need to debug a replication issue"

1. **Check**: [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common problems
2. **Logs**: `kubectl logs -n k8s-replicator-system deployment/k8s-replicator-controller-manager`
3. **Labels**: Verify `replicator.nadundesilva.github.io/object-type: replicated` on source
4. **Namespaces**: Check namespace filtering labels

### "I need to test a specific resource type"

1. **See**: [CONTRIBUTING.md](CONTRIBUTING.md#running-tests-for-specific-resources) for complete instructions on using `TEST_RESOURCES_FILTER_REGEX`

## Critical AI Agent Insights 🧠

### Project-Specific Patterns

- **Bundle Generation**: `make bundle` auto-generates RBAC from kubebuilder comments
- **Test Data Structure**: Each resource type needs its own test data file in `test/utils/testdata/`
- **Controller Registration**: New replicators must be added to `NewReplicators()` function
- **RBAC Comments**: Use `//+kubebuilder:rbac` format in controller files

### Common Pitfalls to Avoid

- **Missing GitHub Actions**: End-to-end test matrix must include new resource types
- **Incomplete RBAC**: Both kubebuilder comments AND generated YAML must be updated
- **Test Coverage**: New resource types need comprehensive test data with realistic scenarios
- **Documentation**: API.md supported resources list is mandatory for new types

**For runtime troubleshooting**: See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for operational issues and debugging.

### File Relationships

- **Replicator Interface**: `controllers/replication/replicator.go` (defines contract)
- **RBAC Generation**: kubebuilder comments → `make bundle` → `config/rbac/role.yaml`
- **Test Integration**: `test/utils/testdata/data.go` includes all resource test data
- **CI/CD Integration**: `.github/workflows/build.yaml` end-to-end matrix tests all resources

**For detailed development patterns and workflows**: See [CONTRIBUTING.md](CONTRIBUTING.md#extending-the-operator-) for comprehensive development guidance.

## Quick Reference 📚

### Key Files

- **Core Interface**: `controllers/replication/replicator.go`
- **Controllers**: `controllers/replication_controller.go`, `controllers/namespace_controller.go`
- **RBAC**: `config/rbac/role.yaml`
- **Tests**: `test/utils/testdata/`
- **CI/CD**: `.github/workflows/build.yaml`

### Key Commands

- **Generate**: `make bundle` (after adding kubebuilder RBAC comments)
- **Test**: `make test`, `make test.e2e`, `make test.unit`
- **Build**: `make build`, `make docker-build`
- **Deploy**: `make install deploy`

### Key Labels

- **Replication**: `replicator.nadundesilva.github.io/object-type: replicated`
- **Replica**: `replicator.nadundesilva.github.io/object-type: replica`
- **Ignore Namespace**: `replicator.nadundesilva.github.io/namespace-type: ignored`

## Maintaining This Guide 📝

**For Contributors**: When making changes to the project, please review and update this `AGENTS.md` file if:

- Adding new documentation files
- Changing file locations or names
- Modifying the core architecture or interfaces
- Adding new development patterns or workflows

This ensures AI agents always have accurate navigation guidance! 🤖✨

**See also**: [CONTRIBUTING.md](CONTRIBUTING.md#ai-agent-usage-policy-) for AI agent usage guidelines and development workflow.

## Content Guidelines 📋

**When updating this file**:

- **Keep it concise**: This is a navigation guide, not detailed documentation
- **Avoid duplication**: Link to authoritative sources instead of repeating content
- **No unnecessary spaces**: Clean formatting with no trailing whitespaces
- **Navigation focus**: Only include information that helps AI agents find the right documentation
- **Link over content**: Always prefer linking to existing documentation rather than adding new content
- **Project philosophy**: This project emphasizes linking over duplication - always direct users to specific documentation rather than repeating information 🔗
