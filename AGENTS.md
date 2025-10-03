# AI Agent Navigation Guide 🤖

> **Quick orientation for AI assistants working with K8s Replicator**

## Project Overview 📋

**K8s Replicator** is a Kubernetes operator that replicates resources across namespaces using labels/annotations. See [README.md](README.md) for project overview and [ARCHITECTURE.md](ARCHITECTURE.md) for technical details.

**Key Concept**: Mark resources with `replicator.nadundesilva.github.io/object-type: replicated` to replicate them across namespaces.

## Quick Start 🚀

**Initial Analysis Sequence:**

1. **Read**: [README.md](README.md) for project overview
2. **Understand**: [ARCHITECTURE.md](ARCHITECTURE.md) for system design
3. **Explore**: `controllers/replication/replicator.go` for core interface
4. **Reference**: [API.md](API.md) for supported resources and labels

**For Changes**: See [CONTRIBUTING.md](CONTRIBUTING.md#development-workflow-) for complete workflow.

## Documentation Map 🗺️

- **[README.md](README.md)** - Project overview, quick start, installation
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Technical architecture, controllers, data flow diagrams
- **[API.md](API.md)** - Complete API reference, supported resources, labels/annotations
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development setup, adding new resource types, testing
- **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Common issues, debugging steps
- **[examples/](examples/)** - Practical examples with setup instructions

## Key Architecture 🏗️

### Controllers

- **Replication Controller**: Watches source resources → replicates to namespaces
- **Namespace Controller**: Watches namespaces → replicates existing source resources to new namespaces
- **No direct communication** between controllers
- **Critical insight**: Namespace Controller replicates ALL existing source resources to new namespaces automatically

### Core Interface

- **Interface**: `controllers/replication/replicator.go` - pure data transformation (no API calls)
- **Pattern**: Each resource type has its own replicator file (e.g., `secret.go`, `serviceaccount.go`)
- **Critical insight**: The `Replicate()` method performs ONLY in-memory data copying - no API calls

## Dos ✅

### Code Analysis

- **Start with**: [ARCHITECTURE.md](ARCHITECTURE.md) for system understanding
- **Check interface**: `controllers/replication/replicator.go` (has comprehensive docstrings)
- **Review controllers**: `*_controller.go` files in `controllers/`
- **Verify RBAC**: `config/rbac/role.yaml`

### Making Changes

- **Read guides**: [CONTRIBUTING.md](CONTRIBUTING.md#adding-new-resource-types) for step-by-step instructions
- **Follow workflow**: [CONTRIBUTING.md](CONTRIBUTING.md#development-workflow-) for complete process
- **Test thoroughly**: [CONTRIBUTING.md](CONTRIBUTING.md#testing) for test commands and structure
- **Run bundle**: Always run `make bundle` after adding kubebuilder RBAC comments
- **Update CI/CD**: Update end-to-end test matrix in `.github/workflows/build.yaml` for new resource types

### Debugging Issues

- **Check troubleshooting**: [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common problems
- **Verify labels**: Use [API.md](API.md) reference for labels/annotations
- **Check logs**: `kubectl logs -n k8s-replicator-system deployment/k8s-replicator-controller-manager`
- **Verify labels**: `replicator.nadundesilva.github.io/object-type: replicated` on source
- **Check namespaces**: Verify namespace filtering labels

### User Assistance

- **Quick start**: Direct to [README.md](README.md)
- **Setup help**: Point to [examples/](examples/) directory
- **Technical questions**: Reference [ARCHITECTURE.md](ARCHITECTURE.md)
- **API questions**: Always link to [API.md](API.md)

## Don'ts ❌

- **Don't duplicate content**: Never repeat information that exists elsewhere - always link to authoritative sources
- **Don't skip RBAC**: Both kubebuilder comments AND generated YAML must be updated
- **Don't forget CI/CD**: End-to-end test matrix must include new resource types
- **Don't skip tests**: New resource types need comprehensive test data with realistic scenarios
- **Don't forget docs**: README.md supported resources list is mandatory for new types
- **Don't break links**: Always update links when moving content between files

## Commands 🔧

### Development

- **Generate**: `make bundle` (after adding kubebuilder RBAC comments)
- **Test**: `make test`, `make test.e2e`, `make test.unit`
- **Build**: `make build`, `make docker-build`
- **Deploy**: `make install deploy`

### Testing Specific Resources

- **Filter tests**: `TEST_RESOURCES_FILTER_REGEX="<ResourceName>" make test.e2e`
- **Resource names**: Use exact names from test data (e.g., `ServiceAccount`, `Secret`, `ConfigMap`, `NetworkPolicy`)

### Debugging

- **Check logs**: `kubectl logs -n k8s-replicator-system deployment/k8s-replicator-controller-manager`
- **Verify permissions**: `kubectl auth can-i create secrets --as=k8s-replicator-system:serviceaccount:k8s-replicator-system:k8s-replicator-controller-manager`

## Key Files 📁

- **Core Interface**: `controllers/replication/replicator.go`
- **Controllers**: `controllers/replication_controller.go`, `controllers/namespace_controller.go`
- **RBAC**: `config/rbac/role.yaml`
- **Tests**: `test/utils/testdata/`
- **CI/CD**: `.github/workflows/build.yaml`

## Common Scenarios 🎯

### "I need to add support for a new resource type"

1. **Read**: [CONTRIBUTING.md](CONTRIBUTING.md#adding-new-resource-types) for complete guide
2. **Create**: New replicator in `controllers/replication/`
3. **Register**: Add to `NewReplicators()` in `controllers/replication/replicator.go`
4. **RBAC**: Add kubebuilder comments in `controllers/replication_controller.go`
5. **Test**: Create test data in `test/utils/testdata/`
6. **Bundle**: Run `make bundle` to generate RBAC
7. **CI/CD**: Update `.github/workflows/build.yaml` end-to-end matrix
8. **Docs**: Update [README.md](README.md#supported-resources) supported resources list

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

## Project Patterns 🔄

### Development Workflow

- **Bundle Generation**: `make bundle` auto-generates RBAC from kubebuilder comments
- **Test Data Structure**: Each resource type needs its own test data file in `test/utils/testdata/`
- **Controller Registration**: New replicators must be added to `NewReplicators()` function
- **RBAC Comments**: Use `//+kubebuilder:rbac` format in controller files

### File Relationships

- **Replicator Interface**: `controllers/replication/replicator.go` (defines contract)
- **RBAC Generation**: kubebuilder comments → `make bundle` → `config/rbac/role.yaml`
- **Test Integration**: `test/utils/testdata/data.go` includes all resource test data
- **CI/CD Integration**: `.github/workflows/build.yaml` end-to-end matrix tests all resources

## Maintenance 📝

**Update this file when:**

- Adding new documentation files
- Changing file locations or names
- Modifying the core architecture or interfaces
- Adding new development patterns or workflows

**See also**: [CONTRIBUTING.md](CONTRIBUTING.md#ai-agent-usage-policy-) for AI agent usage guidelines and development workflow.

## Documentation Philosophy 🎯

**This project follows a strict "no duplication" policy:**

- **Single source of truth**: Each piece of information lives in exactly one place
- **Link, don't copy**: Always link to authoritative sources rather than repeating content
- **Maintain links**: When content moves, update all references immediately
- **Clear hierarchy**: README.md for overview, specialized docs for details
- **AI agent friendly**: Links help agents navigate to the right information quickly

**Examples of this philosophy in action:**

- Supported resources list: Only in README.md, all other files link to it
- API documentation: Detailed in API.md, referenced from other docs
- Development workflows: Complete in CONTRIBUTING.md, linked from AGENTS.md
