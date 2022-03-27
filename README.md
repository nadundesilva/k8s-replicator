# Kubernetes Replicator

[![Main Branch Build](https://github.com/nadundesilva/k8s-replicator/actions/workflows/build-branch.yaml/badge.svg)](https://github.com/nadundesilva/k8s-replicator/actions/workflows/build-branch.yaml)

Replicator for Kubernetes resources across namespaces. This controller was written keeping exptendability in mind. Therefore, it can be extended to any other resource as needed. The following resources are supported by the Kubernetes replicator.

* Secrets

## How to Use

### How to Setup Controller

* Clone this repository.
* Update the configuration (`<REPOSITORY_ROOT>/kustomize/config.yaml`) to match your needs.
* Apply the controller into your cluster.

  ```bash
  kubectl apply -k kustomize
  ```

### How to mark a object to be replicated

Use the following label to mark the object to be replicated.

```properties
replicator.nadundesilva.github.io/object-type=source
```

All objects with the above label will replicated into all namespaces.
