# K8s Replicator - Benchmark Results

These benchmark tests are performed within GitHub Actions, with the tester as well as the Kind K8s cluster sharing the same GitHub action resources. These are only meant as a measure of the relative performance of the Operator over time. When you are running the Operator in a Kubernetes cluster with higher resources allocated to the Kube API server, you can expect much better performance.

## Namespace Creation

This is a benchmark on the duration taken to replicate resources to a set of new namespaces with varying initial and new namespaces counts. The initial namespaces are created beforehand and only the time taken to create the new namespaces and replicate to them are measured for the benchmark.

| Initial Namespace Count | New Namespace Count | Duration |
| -- | -- | -- |
| 0 | 1 | 62.580394ms |
| 0 | 10 | 806.056765ms |
| 0 | 100 | 36.650028397s |
| 0 | 1000 | 6m36.792764865s |
| 1 | 100 | 36.820706206s |
| 10 | 100 | 38.635464841s |
| 100 | 100 | 56.634820848s |
| 1000 | 100 | 3m56.804696605s |

## Resource Creation

This is a benchmark on replicating a new resource to namespaces with varying namespaces counts. The namespaces are created beforehand and only the time to replicate to the new namespaces are measured.

| Namespace Count | Duration |
| -- | -- |
| 1 | 1.047408693s |
| 10 | 2.206127117s |
| 100 | 24.407151076s |
| 1000 | 4m11.420508672s |
