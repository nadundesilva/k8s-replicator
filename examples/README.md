# Examples 📚

Welcome to our examples! 🎯 This directory contains practical examples demonstrating K8s Replicator in real-world scenarios, from basic usage to complex multi-tenant setups.

## Available Examples 🚀

- [**Cert Manager**](./cert-manager) 🔐 - TLS certificate management across microservices

## Quick Start ⚡

**Prerequisites:**

- Kubernetes cluster (v1.20+) ☸️
- kubectl configured and working
- K8s Replicator installed in your cluster
- Sufficient permissions to create resources

**Running the Example:**

```bash
cd examples/cert-manager
cat README.md
bash setup.sh
bash validate.sh
bash clean.sh
```

## Example Categories 📋

**Certificate Management:**

- TLS Certificate Distribution
- Wildcard Certificate Usage
- Certificate Rotation

**Common Use Cases:**

- Multi-Tenant Applications
- Microservices Architecture
- Security and Compliance
- Configuration Management

## Best Practices 💡

**Resource Organization:**

- Use meaningful names for resources and namespaces
- Apply consistent labeling for easy identification
- Document resource purpose in annotations
- Follow naming conventions for your organization

**Security Considerations:**

- Validate resource content before replication
- Use proper RBAC to restrict access
- Encrypt sensitive data in secrets
- Audit replication operations regularly

**Performance Optimization:**

- Batch operations when possible
- Monitor resource usage and limits
- Use appropriate resource quotas
- Implement proper error handling

## Troubleshooting 🔧

**Example Setup Fails:**

```bash
kubectl get pods -n k8s-replicator-system
kubectl auth can-i create secrets --as=k8s-replicator-system:serviceaccount:k8s-replicator-system:k8s-replicator-controller-manager
```

**Resources Not Replicating:**

```bash
kubectl get secret my-secret -o yaml | grep replicator.nadundesilva.github.io
kubectl get namespace my-namespace -o yaml | grep replicator.nadundesilva.github.io
```

**Permission Denied Errors:**

```bash
kubectl get clusterrole k8s-replicator-manager-role -o yaml
kubectl get clusterrolebinding k8s-replicator-manager-rolebinding -o yaml
```

## Contributing Examples 🤝

**Create Example Directory:**

```bash
mkdir examples/your-example
cd examples/your-example
```

**Required Files:**

- `README.md` - Example documentation
- `setup.sh` - Setup script
- `validate.sh` - Validation script
- `clean.sh` - Cleanup script
- Resource YAML files

**Example Structure:**

```text
your-example/
├── README.md          # Documentation
├── setup.sh          # Setup script
├── validate.sh       # Validation script
├── clean.sh          # Cleanup script
├── resources/        # Resource definitions
└── kustomization.yaml # Kustomize configuration
```

## Support 💬

- [GitHub Issues](https://github.com/nadundesilva/k8s-replicator/issues/new)
- [GitHub Discussions](https://github.com/nadundesilva/k8s-replicator/discussions)
- [Full Documentation](../README.md)

---

**Happy Replicating!** 🚀
