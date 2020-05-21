# Zilliqa Exporter

A daemon that exports metrics of Zilliqa node to Prometheus

## Common Labels:
| Name         | Description                                          |
|:-------------|:-----------------------------------------------------|
| type         | type of node. (normal,lookup,newlookup,level2lookup) |
| pod_name     | node's pod name                                      |
| cluster_name | name of genet cluster                                |
| network_name | name of zilliqa network                              |
| public_ip    | public IP address                                    |
| local_ip     | local private IP address                             |



## Metrics & Source

### Scheduled Collector

| Metric           | Description                                      | Source                                                | Method                | Period | Mainnet Only |
|:-----------------|:-------------------------------------------------|:------------------------------------------------------|:----------------------|:-------|:-------------|
| node_info        | Node Information of zilliqa and host environment | CMD Options & EnvVars & AWS Metadata & Zilliqa Binary | -                     | 30m    | No           |
| ud_state_size    | State data size of unstoppable domain contract   | api                                                   | GetSmartContractState | 1h     | Yes          |
| ud_state_entries | State records of unstoppable domain contract     | api                                                   | GetSmartContractState | 1h     | Yes          |

### Instant Collector

| Metric        | Description                                        | Source      | Method                                  |
|:--------------|:---------------------------------------------------|:------------|:----------------------------------------|
| epoch         | Current TX block number of the node                | api & admin | GetBlockchainInfo & GetCurrentMiniEpoch |
| ds_epoch      | Current DS block number of the node                | api & admin | GetBlockchainInfo & GetCurrentDSEpoch   |
| difficulty    | The minimum shard difficulty of the previous block | api & admin | GetPrevDifficulty                       |
| ds_difficulty | The minimum DS difficulty of the previous block    | api & admin | GetPrevDSDifficulty                     |
| network_id    | network ID of current zilliqa network              | api         | GetNetworkId                            |
| node_type     | Zilliqa network node type                          | admin       | GetNodeType                             |
| shard_num     | Shard number of current node                       | admin       | GetNodeType                             |
| node_state    | Node consensus state                               | admin       | GetNodeState                            |

### Psutil Collector

Get process information using go-psutil

| Label | Description               |
|:------|:--------------------------|
| pid   | process ID                |
| cwd   | current working directory |

| Metric           | Description                                         | unit  |
|:-----------------|:----------------------------------------------------|:------|
| up               | If zilliqa process is running                       | -     |
| synctype         | Synctype from zilliqa commandline option            | -     |
| node_uptime      | Uptime of zilliqa node                              | ms    |
| connection_count | Network Connection count of zilliqa process         | -     |
| thread_count     | Thread count of zilliqa process                     | -     |
| fd_count         | Opened file descriptor count of zilliqa process     | -     |
| storage_total    | Total capacity of zilliqa persistence storage (cwd) | bytes |
| storage_used     | Used space of zilliqa persistence storage (cwd)     | bytes |

<!--| cpu_percent      | CPU usage percent of zilliqa process                | -     |-->
<!--| mem_percent      | Memory usage percent of zilliqa process             | -     |-->
