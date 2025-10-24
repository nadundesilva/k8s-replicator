# K8s Replicator - Benchmark Results

These benchmark tests are performed within GitHub Actions, with the tester as well as the Kind K8s cluster sharing the same GitHub action resources. These are only meant as a measure of the relative performance of the Operator over time. When you are running the Operator in a Kubernetes cluster with higher resources allocated to the Kube API server, you can expect much better performance.

## Namespace Creation

This is a benchmark on the duration taken to replicate resources to a set of new namespaces with varying initial and new namespaces counts. The initial namespaces are created beforehand and only the time taken to create the new namespaces and replicate to them are measured for the benchmark.

| Initial Namespace Count | New Namespace Count | Duration |
| -- | -- | -- |
| 0 | 1 | 54.368824ms |
| 0 | 10 | 827.092195ms |
| 0 | 100 | 20.482477055s |
| 0 | 1000 | 3m37.823926948s |
| 1 | 100 | 20.700008856s |
| 10 | 100 | 22.514077477s |
| 100 | 100 | 40.52703045s |
| 1000 | 100 | 3m40.938433551s |

## Resource Creation

This is a benchmark on replicating a new resource to namespaces with varying namespaces counts. The namespaces are created beforehand and only the time to replicate to the new namespaces are measured.

| Namespace Count | Duration |
| -- | -- |
| 1 | 58.074283ms |
| 10 | 1.062229228s |
| 100 | 20.269171609s |
| 1000 | 3m21.792481153s |
