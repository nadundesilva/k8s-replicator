# K8s Replicator - Benchmark Results

These benchmark tests are performed within GitHub Actions, with the tester as well as the Kind K8s cluster sharing the same GitHub Action resources. These are only meant as a measure of the relative performance of the Operator over time. When you are running the Operator in a Kubernetes cluster with higher resources allocated to the Kube API server, you can expect much better performance.

## Namespace Creation

This is a benchmark on the duration taken to replicate resources to a set of new namespaces with varying initial and new namespaces counts. The initial namespaces are created beforehand and only the time taken to create the new namespaces and replicate to them are measured for the benchmark.

| Initial Namespace Count | New Namespace Count | Duration        |
| ----------------------- | ------------------- | --------------- |
| 0                       | 1                   | 16.654617ms     |
| 0                       | 10                  | 699.238624ms    |
| 0                       | 100                 | 36.615823017s   |
| 0                       | 1000                | 6m36.687582862s |
| 1                       | 100                 | 36.819997424s   |
| 10                      | 100                 | 38.61690963s    |
| 100                     | 100                 | 56.634983989s   |
| 1000                    | 100                 | 3m56.730384949s |

## Resource Creation

This is a benchmark on replicating a new resource to namespaces with varying namespaces counts. The namespaces are created beforehand and only the time to replicate to the new namespaces are measured.

| Namespace Count | Duration        |
| --------------- | --------------- |
| 1               | 22.06732ms      |
| 10              | 2.402786763s    |
| 100             | 24.003125332s   |
| 1000            | 4m11.802919827s |
