# Kubernetes Replicator

Replicator for Kubernetes resources across namespaces. This controller was written keeping exptendability in mind. Therefore, it can be extended to any other resource as needed. The following resources are supported by the Kubernetes replicator.

* Secrets

## How to Use

* Clone this repository.
* Update the configuration (`<REPOSITORY_ROOT>/kustomize/config.yaml`) to match your needs.
* Apply the controller into your cluster.
  ```bash
  kubectl apply -k kustomize
  ```
