# K8s Replicator - Benchmark Results

These benchmark tests are performed within GitHub Actions, with the tester as well as the Kind K8s cluster sharing the same GitHub action resources. These are only meant as a measure of the relative performance of the Operator over time. When you are running the Operator in a Kubernetes cluster with higher resources allocated to the Kube API server, you can expect much better performance.

## Namespace Creation

This is a benchmark on the duration taken to replicate resources to a set of new namespaces with varying initial and new namespaces counts. The initial namespaces are created beforehand and only the time taken to create the new namespaces and replicate to them are measured for the benchmark.

| Initial Namespace Count | New Namespace Count | Duration |
| -- | -- | -- |
| 0 | 1 | 48.21516ms |
| 0 | 10 | 819.575286ms |
| 0 | 100 | 36.624389594s |
| 0 | 1000 | 6m36.763401916s |
| 1 | 100 | 36.830349127s |
| 10 | 100 | 38.632514298s |
| 100 | 100 | 56.637396105s |
| 1000 | 100 | 3m56.765937712s |

## Resource Creation

This is a benchmark on replicating a new resource to namespaces with varying namespaces counts. The namespaces are created beforehand and only the time to replicate to the new namespaces are measured.

| Namespace Count | Duration |
| -- | -- |
| 1 | 1.047526764s |
| 10 | 2.204460298s |
| 100 | 23.404635028s |
| 1000 | 4m5.408961946s |
