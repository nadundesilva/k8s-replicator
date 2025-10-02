# AI Agent Navigation Guide 🤖

> **Quick orientation for AI assistants working with K8s Replicator**

## Project Type & Context 📋

**K8s Replicator** is a Kubernetes operator (Go + controller-runtime) that replicates resources across namespaces using labels/annotations.

## Essential Documentation Map 🗺️

### Start Here First

- **[README.md](README.md)** - Project overview, quick start, installation
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Technical architecture, controllers, data flow diagrams

### For Code Analysis & Development

- **[API.md](API.md)** - Complete API reference, supported resources (definitive list), labels/annotations
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development setup, adding new resource types, testing
- **Core implementation**: `internal/controller/` directory
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

### Core Interface Pattern

- **Complete interface definition**: `internal/controller/replication/replicator.go` (with full Go docstrings)
- **Implementation examples**: `internal/controller/replication/*.go` files

### Labels/Annotations System

- **Complete reference**: [API.md](API.md) - all labels, annotations, and values

## AI Agent Guidelines 🎯

### When Analyzing Code

1. **Start with**: [ARCHITECTURE.md](ARCHITECTURE.md) for system understanding
2. **Interface details**: `internal/controller/replication/replicator.go` (has comprehensive docstrings)
3. **Controller logic**: `*_controller.go` files in `internal/controller/`
4. **RBAC requirements**: `config/rbac/role.yaml`

### When Making Changes

1. **Follow patterns**: See [CONTRIBUTING.md](CONTRIBUTING.md) for adding new resource types
2. **Update documentation**: [API.md](API.md) supported resources section is mandatory
3. **Test requirements**: See `test/` directory structure
4. **RBAC updates**: Required for new resource types

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

## File Priority for AI Agents 📁

### High Priority (Read First)

- `README.md` - Project overview
- `ARCHITECTURE.md` - System design
- `internal/controller/replication/replicator.go` - Core interface
- `API.md` - Complete API reference

### Medium Priority (Context Dependent)

- `CONTRIBUTING.md` - When helping with development
- `TROUBLESHOOTING.md` - When debugging issues
- `examples/` - When explaining usage
- `config/rbac/` - When dealing with permissions

### Low Priority (Reference Only)

- `CHANGELOG.md`, `SECURITY.md`, `ADOPTERS.md` - Historical/meta information
- `test/` - When writing tests
- `hack/`, `installers/` - Deployment specifics

## Quick Facts ⚡

- **Deployment details**: [README.md](README.md) installation section
- **Supported resources**: [API.md](API.md#supported-resources) (definitive list)
- **Extension guide**: [CONTRIBUTING.md](CONTRIBUTING.md) adding new resource types
- **Release info**: [CHANGELOG.md](CHANGELOG.md)

## Maintaining This Guide 📝

**For Contributors**: When making changes to the project, please review and update this `AGENTS.md` file if:

- Adding new documentation files
- Changing file locations or names
- Modifying the core architecture or interfaces
- Adding new development patterns or workflows
- Updating key configuration or deployment processes

This ensures AI agents always have accurate navigation guidance! 🤖✨

## Content Guidelines 📋

**When updating this file**:

- **Keep it concise**: This is a navigation guide, not detailed documentation
- **Avoid duplication**: Link to authoritative sources instead of repeating content
- **No unnecessary spaces**: Clean formatting with no trailing whitespaces
- **Navigation focus**: Only include information that helps AI agents find the right documentation
- **Link over content**: Always prefer linking to existing documentation rather than adding new content
- **Project philosophy**: This project emphasizes linking over duplication - always direct users to specific documentation rather than repeating information 🔗
