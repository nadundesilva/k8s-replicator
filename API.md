# API Documentation 📚

Welcome! 👋 This document provides a comprehensive API reference for K8s Replicator, including the Replicator interface, configuration options, and extension points.

## Replicator Interface 🔌

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

### Method Documentation

**GetKind()** - Returns the Kubernetes resource kind that this replicator handles (e.g., "Secret", "ConfigMap", "NetworkPolicy", or any Kubernetes resource)

**AddToScheme()** - Registers the resource type with the Kubernetes scheme for proper serialization/deserialization

**EmptyObject()** - Returns an empty instance of the resource type for use in API operations

**EmptyObjectList()** - Returns an empty list of the resource type for use in list operations

**ObjectListToArray()** - Converts a client.ObjectList to a slice of client.Object for easier processing

**Replicate()** - Copies all the data of one K8s object that should be replicated to another K8s object

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

- `ResourceNotFound`: Source resource not found - Verify resource exists and has correct labels
- `NamespaceNotFound`: Target namespace not found - Create namespace or check filtering
- `PermissionDenied`: Insufficient RBAC permissions - Check and update RBAC configuration
- `ResourceConflict`: Resource already exists - Delete conflicting resource or update logic

---

For more examples, see [Examples](examples/) directory. 🚀

Happy coding! 💻✨
