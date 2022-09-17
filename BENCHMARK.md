# K8s Replicator - Benchmark Results

## Namespace Creation

This is a benchmark on the duration taken to replicate resources to a set of new namespaces with varying initial and new namespaces counts. The initial namespaces are created beforehand and only the time taken to create the new namespaces and replicate to them are measured for the benchmark.

| Initial Namespace Count | New Namespace Count | Duration |
| -- | -- | -- |
| 0 | 1 | 22.724364ms |
| 0 | 10 | 4.691715001s |
| 0 | 100 | 58.621989867s |
| 0 | 1000 | 10m8.718282655s |
| 1 | 100 | 57.824094674s |
| 10 | 100 | 58.626700309s |
| 100 | 100 | 1m4.436625486s |
| 1000 | 100 | 3m59.32179761s |

## Resource Creation

This is a benchmark on replicating a new resource to namespaces with varying namespaces counts. The namespaces are created beforehand and only the time to replicate to the new namespaces are measured.

| Namespace Count | Duration |
| -- | -- |
| 1 | 1.045284328s |
| 10 | 3.808940571s |
| 100 | 41.20444918s |
| 1000 | 7m4.204291703s |
