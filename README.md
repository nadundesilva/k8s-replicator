# K8s Replicator

[![Main Branch Build](https://github.com/nadundesilva/k8s-replicator/actions/workflows/branch-build.yaml/badge.svg)](https://github.com/nadundesilva/k8s-replicator/actions/workflows/branch-build.yaml)
[![Vulnerabilities Scan](https://github.com/nadundesilva/k8s-replicator/actions/workflows/vulnerabilities-scan.yaml/badge.svg)](https://github.com/nadundesilva/k8s-replicator/actions/workflows/vulnerabilities-scan.yaml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[![Release](https://img.shields.io/github/release/nadundesilva/k8s-replicator.svg?style=flat-square)](https://github.com/nadundesilva/k8s-replicator/releases/latest)
[![Docker Image](https://img.shields.io/docker/image-size/nadunrds/k8s-replicator/latest?style=flat-square)](https://hub.docker.com/r/nadunrds/k8s-replicator)
[![Docker Pulls](https://img.shields.io/docker/pulls/nadunrds/k8s-replicator?style=flat-square)](https://hub.docker.com/r/nadunrds/k8s-replicator)

Replicator supports copying kubernetes objects across namespaces. This controller was written keeping extensibility and [performance](./BENCHMARK.md) in mind. Therefore, it can be extended to any other resource as needed. The following resources are supported by the K8s replicator.

- Secrets
- Config Maps
- Network Policies

## How to Use

### Prerequisites

The following tools are expected to be installed and ready.

- Kubectl
- Operator SDK

The following tools can be either installed on your own or let the installation scripts handle it.

- OLM to be installed in the cluster
  OLM can be installed using the [operator-sdk](https://sdk.operatorframework.io/docs/installation/)
  ```bash
  operator-sdk olm install
  ```

### How to Setup Operator

#### Quickstart

Run the following command to apply the controller to your cluster. The `<VERSION>` should be replaced with the release version
to be used (eg:- `0.1.0`) and kubectl CLI should be configured pointing to the cluster in which the controller needs to be started.

```bash
curl -L https://raw.githubusercontent.com/nadundesilva/k8s-replicator/main/installers/install.sh | bash -s <VERSION>
```

#### Manual Installation

- Make sure all the pre-requisites are installed (including the dependencies which are normally installed by the installation scripts)
- Install the Operator Bundle using the Operator SDK. The `<VERSION>` should be replaced with the release version
  to be used (eg:- `0.1.0`) and kubectl CLI should be configured pointing to the cluster in which the controller needs to be started.
  ```bash
  operator-sdk run bundle docker.io/nadunrds/k8s-replicator-bundle:<VERSION>
  ```

### How to mark a object to be replicated

Use the following label to mark the object to be replicated.

```properties
k8s-replicator.nadundesilva.github.io/object-type=replicated
```

All objects with the above label will be replicated into all namespaces. You can configure the replication behavior using the below annotations (You still need to add the above label to mark them as objects that should be replicated).

* `k8s-replicator.nadundesilva.github.io/replicated-to`

  By default, if this annotation is not present, the replicated objects will be replicated to all namespaces that are
  not ignored. By using this annotation, you can specify a list of comma-separated namespaces to which the object will
  be replicated. If the value is an empty string (`""`), the object will not be replicated at all.

* `k8s-replicator.nadundesilva.github.io/replicated-to-matching`

  This has the same behavior as the `k8s-replicator.nadundesilva.github.io/replicated-to` annotation. However, the value
  of this annotation should be a valid labels selector
  (either [equality-based](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#equality-based-requirement)
  or [set-based](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#set-based-requirement)).

* `k8s-replicator.nadundesilva.github.io/strip-labels`

  This annotation can be used to strip any labels that you wish to not be present in the replicas. By specifying a
  comma-separated list of label keys, you can remove those labels from the replicas. If the label that you specified
  is not present in the source object, it will not have any impact.  The values in this annotation cannot include
  any of the labels owned by this operator as they will be added and removed properly as deemed fit by the Operator.

* `k8s-replicator.nadundesilva.github.io/strip-annotations`

  Similar to the `k8s-replicator.nadundesilva.github.io/strip-labels` annotation, this can be used to remove any
  unwanted annotations. A comma-separated list of annotation keys should be specified as the value and if they are
  not present in the source object, it will not take any action for the missing annotations. The values in this
  annotation cannot include any of the annotations owned by this operator as they will be added and removed properly
  as deemed fit by the Operator.

* `k8s-replicator.nadundesilva.github.io/preserve-owner-reference`

  By default, the Operator will not copy any owner references in the source object. By using this annotation,
  you can copy the owner references as well into the replicas.

#### Ignored namespaces

The following namespaces are ignored by default.

* The namespace in which controller resides
* Namespaces with the name starting with `kube-` prefix
* Namespaces with the ignored label
  ```properties
  k8s-replicator.nadundesilva.github.io/namespace-type=ignored
  ```

If you want to override this behavior and specifically replicate to a namespace, add the following label

```properties
k8s-replicator.nadundesilva.github.io/namespace-type=managed
```

`k8s-replicator.nadundesilva.github.io/namespace-type` label has higher priority over all other labels. If a namespace is ignored, any objects of type replicated in the namespace will not be replicated, and no replicas will be created in them.

### Examples

Examples for the CRDs used by the Operator can be found in the [samples](./config/samples) directory.

### Additional labels/annotations used by the controller

The folloing labels are used by the controller to track the replication of objects.

* The following label with the value `replica` is used to mark the replicated objects.
  ```properties
  k8s-replicator.nadundesilva.github.io/object-type=replica
  ```
* The following annotation is used to store a replica's source namespace.
  ```properties
  k8s-replicator.nadundesilva.github.io/source-namespace=<namespace>
  ```

### How to Cleanup Operator

#### Quick Remove

Run the following command to remove the controller from your cluster. Kubectl CLI should be configured pointing to the cluster in which the controller needs to be started.

```bash
curl -L https://raw.githubusercontent.com/nadundesilva/k8s-replicator/main/installers/uninstall.sh | bash -s
```

#### Manual Removal

Remove the controller from your cluster by running the following command.

```bash
operator-sdk cleanup k8s-replicator
```

## How to Extend

This Operator is created with extensibility in mind. To support this, a common interface `Replicator` was introduced.

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

The K8s Replicator core uses the methods defined in this interface to get, list, and replicate resources to namespaces. The methods are carefully chosen to ensure that the minimum set of functionalities are defined for each resource separately keeping most of the logic reusable in the Operator core.

You can check the [existing implementations](./internal/controller/replication/) of `Replicator` to get an idea of what needs to be done. However, you need to build the Operator from the source to get the new `Replicator` up and running. That being said, if you wish to contribute new resource replicators, you are most welcome.

## Support

:grey_question: If you need support or have a question about the K8s Replicator, reach out through [Discussions](https://github.com/nadundesilva/k8s-replicator/discussions).

:bug: If you have found a bug and would like to get it fixed, try opening a [Bug Report](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FBug&template=bug-report.md).

:bulb: If you have a new idea or want to get a new feature or improvement added to the K8s Replicator, try creating a [Feature Request](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FFeature&template=feature-request.md).
