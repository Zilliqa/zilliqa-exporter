# Zilliqa Exporter

A daemon that exports metrics of Zilliqa node to Prometheus

# Metrics

Lables:
| Name     | Description                          |
| -------- | ------------------------------------ |
| pod_name | node's pod name                      |
| index    | the index of node                    |
| ip       | the IP address of node               |
| type     | normal,lookup,newlookup,level2lookup |
| is_guard |                                      |
| network  |                                      |



| Metric                      | Description              | Value   |
| --------------------------- | ------------------------ | ------- |
| epoch                       | the epoch number of node | integer |
| last_tx_block_num           |                          |         |
| last_tx_block_num_timestamp |                          |         |
|                             |