# K8s Replicator - Benchmark Results

## Namespace Creation

This is a benchmark on the duration taken to replicate resources to a set of new namespaces with varying initial and new namespaces counts. The initial namespaces are created beforehand and only the time taken to create the new namespaces and replicate to them are measured for the benchmark.

| Initial Namespace Count | New Namespace Count | Duration |
| -- | -- | -- |
| 0 | 1 | 33.42301ms |
| 0 | 10 | 4.706934676s |
| 0 | 100 | 58.628938201s |
| 0 | 1000 | 10m8.772157407s |
| 1 | 100 | 57.834245965s |
| 10 | 100 | 58.624531031s |
| 100 | 100 | 1m3.431816922s |
| 1000 | 100 | 3m59.559109427s |

## Resource Creation

This is a benchmark on replicating a new resource to namespaces with varying namespaces counts. The namespaces are created beforehand and only the time to replicate to the new namespaces are measured.

| Namespace Count | Duration |
| -- | -- |
| 1 | 1.055666516s |
| 10 | 3.81482857s |
| 100 | 41.204071538s |
| 1000 | 7m0.609495945s |
