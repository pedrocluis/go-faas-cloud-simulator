
# go-faas-cloud-simulator

Cloud FaaS Workload Simulator written in Go using the [Azure Functions Trace from 2019](https://github.com/Azure/AzurePublicDataset/blob/master/AzureFunctionsDataset2019.md).
## Installation

Build the project

```bash
  $ ./build.sh
```
## How to run

To run the project:

```bash
  $ ./sim
```

Options:

- -cold_lat int: Cold start latency (default: 250ms)
- -disk int: Disk cache capacity (default 250000MB)
- -input string: Input file
- -keep_alive int: Keep alive window (default 5 minutes)
- -nodes int: Number of nodes (default 80)
- -ram int: RAM capacity (default 32000MB)
- -read_speed float: Disk read bandwidth (default 10GB/s)
- -write_speed float: Disk write bandwidth (default 10GB/s)
- -stats string: Stats output file
- -threads int: Number of threads (default 4)


