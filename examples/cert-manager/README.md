# Cert Manager Example

This example demonstrates how the K8s Replicator can be used along with the [Cert Manager](https://cert-manager.io/), to share a common TLS certificate across multiple microservices.
Since the Kubernetes security model does not allow pointing to a TLS secret across namespaces, if you have TLS certificates shared across multiple different microservices, the secret would need to be copied across namespaces.
While we can copy this manually, using the K8s Replicator would help automate copying as well as rotation of certificates as the K8s Replicator will automatically propagate updates to resources also into all the replicated namespaces.

This examples contains three VSCode editors with a shared common wildcard TLS secret (domain: `*.vscode.local`). The three editors are exposed over three different hostnames using the same wildcard TLS secret.
- editor-01.vscode.local
- editor-02.vscode.local
- editor-03.vscode.local

Cert Manager is used to issue a certificate from a common self signed self certificate.
If you wish to play around with this example, you can update the way the seccret is issued to any technique and use the K8s Replicator along with it.
