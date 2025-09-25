# Cert Manager Example 🔐

Welcome to the Cert Manager example! 🔐 This demonstrates how to use K8s Replicator with [Cert Manager](https://cert-manager.io/) to share TLS certificates across multiple microservices.

**Problem:** Kubernetes security model doesn't allow pointing to TLS secrets across namespaces, so certificates need to be copied manually.

**Solution:** K8s Replicator automates copying and rotation of certificates, automatically propagating updates to all replicated namespaces.

**Example:** Three Visual Studio Code editors with shared wildcard TLS secret (`*.vscode.local`):

- editor-01.vscode.local
- editor-02.vscode.local
- editor-03.vscode.local

Cert Manager issues certificate from a common self-signed certificate. You can modify the certificate issuance method and use K8s Replicator with any technique.

Happy certificate management! 🔐✨
