# Examples 📚

Welcome to our examples! 🎯 This directory contains practical examples demonstrating K8s Replicator in real-world scenarios, from basic usage to complex multi-tenant setups.

## Available Examples 🚀

Each example includes complete setup, validation, and cleanup scripts along with detailed documentation.

- [**Cert Manager**](./cert-manager) 🔐 - TLS certificate management across microservices

## Quick Start ⚡

**Prerequisites:**

See the main [Installation Guide](../README.md#quick-start-) for K8s Replicator setup requirements.

**Running Any Example:**

1. Navigate to the example directory
2. Read the README.md for specific instructions
3. Run the setup, validation, and cleanup scripts

## Example Categories 📋

**Available Categories:**

- **Certificate Management** - TLS certificate distribution and rotation
- **Configuration Management** - Sharing configuration across namespaces
- **Security Policies** - Consistent security policy application
- **Multi-Tenant Applications** - Resource sharing in multi-tenant environments

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

For comprehensive troubleshooting guidance, see the [Troubleshooting Guide](../TROUBLESHOOTING.md).

**Example-Specific Issues:**
- Check that K8s Replicator is properly installed and running
- Verify example-specific resource configurations
- Ensure proper permissions for example resources

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

For support options, see the main [Support section](../README.md#support-) in the project documentation.

---

**Happy Replicating!** 🚀
