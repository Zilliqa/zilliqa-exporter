# Zilliqa Exporter

A daemon that exports metrics of Zilliqa node as [Open Metrics](https://github.com/OpenObservability/OpenMetrics) Format

## Common Labels:

| Name         | Description                                                                                                        |
| :----------- | :----------------------------------------------------------------------------------------------------------------- |
| type         | type of node, can auto detect from pod name or zilliqad commandline params. (normal,lookup,newlookup,level2lookup) |
| pod_name     | node's pod name, from env "POD_NAME" or "Z7A_POD_NAME"                                                             |
| pod_ip       | IP of the pod, from env "POD_IP", "Z7A_POD_IP"                                                                     |
| cluster_name | name of genet cluster, from env "CLUSTER_NAME"                                                                     |
| network_name | name of zilliqa network, from env "Z7A_TESTNET_NAME", "TESTNET_NAME", "NETWORK_NAME"                               |
| public_ip    | public IP address, from AWS metadata                                                                               |
| local_ip     | local private IP address, from AWS metadata                                                                        |


## Metrics & Source

### Constant Collector

Collect constant info from environment variables, AWS metadata and Zilliqa commandline.

Only collect once when exporter starts.

| Metric    | Description                                      | Source                                                |
| :-------- | :----------------------------------------------- | :---------------------------------------------------- |
| node_info | Node Information of zilliqa and host environment | CMD Options & EnvVars & AWS Metadata & Zilliqa Binary |


### API Collector

Collect info from zilliqa node's JSONRPC API server
Only for Lookup, Seed, Seed-apipub(Level2Lookup)

| Metric                   | Description                                         | Method              | Additional Labels |
| :----------------------- | :-------------------------------------------------- | :------------------ | :---------------- |
| api_server_up            | JsonRPC API server up and running                   | -                   | endpoint          |
| epoch                    | Current TX block number of the node                 | GetBlockchainInfo   | -                 |
| ds_epoch                 | Current DS block number of the node                 | GetBlockchainInfo   | -                 |
| transaction_rate         | Current transaction rate                            | GetBlockchainInfo   | -                 |
| tx_block_rate            | Current TX block rate                               | GetBlockchainInfo   | -                 |
| ds_block_rate            | Current DS block rate                               | GetBlockchainInfo   | -                 |
| num_peers                | Peers count                                         | GetBlockchainInfo   | -                 |
| sharding_peers           | Peers count of every shard                          | GetBlockchainInfo   | index (of shard)  |
| num_txns_tx_epoch        | numTxnsTxEpoch                                      | GetBlockchainInfo   | -                 |
| num_txns_ds_epoch        | numTxnsDSEpoch                                      | GetBlockchainInfo   | -                 |
| num_transactions         | Transactions count                                  | GetBlockchainInfo   | -                 |
| difficulty               | The minimum shard difficulty of the previous block  | GetPrevDifficulty   | -                 |
| ds_difficulty            | The minimum DS difficulty of the previous block     | GetPrevDSDifficulty | -                 |
| network_id               | Network ID of current zilliqa network               | GetNetworkId        | -                 |
| latest_txblock_timestamp | The timestamp of the latest tx block (milliseconds) | GetLatestTxBlock    | -                 |

~~Mainnet Only Metrics (scheduled):~~

| Metric           | Description                                    | Method                | Period |
| :--------------- | :--------------------------------------------- | :-------------------- | :----- |
| ud_state_size    | State data size of unstoppable domain contract | GetSmartContractState | 1h     |
| ud_state_entries | State records of unstoppable domain contract   | GetSmartContractState | 1h     |

### Admin Collector

Collect info from zilliqa node's Admin API server (Status Server)

| Metric          | Description                         | Method      | Additional Labels                  |
| :-------------- | :---------------------------------- | :---------- | :--------------------------------- |
| admin_server_up | Admin JsonRPC server up and running | -           | endpoint                           |
| node_type       | Zilliqa network node type           | GetNodeType | text (representative of node type) |

Only for Shard Node:

| Metric        | Description                                        | Method              | Additional Labels |
| :------------ | :------------------------------------------------- | :------------------ | :---------------- |
| shard_id      | Shard ID of the shard of current node              | GetNodeType         | -                 |
| epoch         | Current TX block number of the node                | GetBlockchainInfo   | -                 |
| ds_epoch      | Current DS block number of the node                | GetBlockchainInfo   | -                 |
| difficulty    | The minimum shard difficulty of the previous block | GetPrevDifficulty   | -                 |
| ds_difficulty | The minimum DS difficulty of the previous block    | GetPrevDSDifficulty | -                 |


Not implemented Yet:

| Metric     | Description | Method       | Additional Labels |
| :--------- | :---------- | :----------- | :---------------- |
| node_state | Node state  | GetNodeState | -                 |

### ProcessInfo Collector

Get running process information

| Label         | Description               |
| :------------ | :------------------------ |
| process_name  | Process Name              |
| pid           | Process ID                |
| cwd           | Current working directory |

| Metric                  | Description                                         | unit         | Additional Labels                  |
| :---------------------- | :-------------------------------------------------- | :----------- | :--------------------------------- |
| zilliqa_process_running | If zilliqa process is running                       | -            |                                    |
| synctype                | Synctype from zilliqa commandline option            | -            |                                    |
| nodetype                | Nodetype from zilliqa commandline option            | -            | text (representative of node type) |
| nodeindex               | Nodeindex from zilliqa commandline option           | -            |                                    |
| node_uptime             | Uptime of zilliqa node (unix timestamp)             | milliseconds |                                    |
| connection_count        | Network Connection count of zilliqa process         | -            | local_port, status                 |
| thread_count            | Thread count of zilliqa process                     | -            |                                    |
| fd_count                | Opened file descriptor count of zilliqa process     | -            |                                    |
| storage_total           | Total capacity of zilliqa persistence storage (cwd) | bytes        |                                    |
| storage_used            | Used space of zilliqa persistence storage (cwd)     | bytes        |                                    |
