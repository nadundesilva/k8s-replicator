# API Documentation 📚

Welcome! 👋 This document provides a comprehensive API reference for K8s Replicator, including the Replicator interface, configuration options, and extension points.

## Replicator Interface 🔌

The core of K8s Replicator's extensible architecture is the `Replicator` interface, which defines how different Kubernetes resource types are replicated across namespaces.

**📖 Interface Definition & Documentation:**

- **Source Code**: [`controllers/replication/replicator.go`](controllers/replication/replicator.go)
- **Go Documentation**: See the source file for comprehensive method documentation and usage examples

**🔧 Key Features:**

- **Extensible Design**: Easy to add new resource types without core changes
- **Clean Architecture**: Pure data transformation separate from API operations
- **Type Safety**: Strongly typed interface for reliable operations
- **Resource Agnostic**: Works with any Kubernetes resource
- **Performance Optimized**: In-memory operations with efficient API usage

## Configuration ⚙️

The operator uses standard controller-runtime configuration:

- **Logging**: Configured via `-zap-log-level` flag (default: `1`)
- **Leader Election**: Enabled via `--leader-elect` flag
- **Metrics**: Available on port `:8080`
- **Health Probes**: Available on port `:8081`

## Labels and Annotations 🏷️

### Replication Labels

**`replicator.nadundesilva.github.io/object-type`**

- `replicated`: Marks a resource for replication
- `replica`: Marks a replicated resource

**`replicator.nadundesilva.github.io/namespace-type`**

- `ignored`: Namespace is ignored for replication
- `managed`: Namespace is explicitly managed (overrides ignore)

### Replication Annotations

**`replicator.nadundesilva.github.io/source-namespace`**

- Stores the source namespace of a replicated resource

## Supported Resources 🔧

**Currently Supported Resource Types:**

- **Secrets** 🔐
- **ConfigMaps** 📄
- **NetworkPolicies** 🛡️

The system was designed with an extensible architecture that allows easy addition of new resource types as needed.

**Need support for a different resource type?** See the [Contributing Guide](CONTRIBUTING.md#extending-the-operator) for implementation instructions.

## Examples 💡

**Secret Replication:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  labels:
    replicator.nadundesilva.github.io/object-type: replicated
type: Opaque
data:
  username: YWRtaW4=
  password: cGFzc3dvcmQ=
```

**Namespace Filtering:**

```yaml
# Ignore namespace
apiVersion: v1
kind: Namespace
metadata:
  name: ignored-namespace
  labels:
    replicator.nadundesilva.github.io/namespace-type: ignored
```

## Error Handling 🚨

**Common Errors:**

- `ResourceNotFound`: Source resource not found - Double-check the resource exists and has correct labels
- `NamespaceNotFound`: Target namespace not found - Create the namespace or check your filtering
- `PermissionDenied`: Insufficient RBAC permissions - Review and update your RBAC configuration
- `ResourceConflict`: Resource already exists - Delete conflicting resource or update logic

---

For more examples, see [Examples](examples/) directory. 🚀

Happy coding! 💻✨
