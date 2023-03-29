# sei-cosmos-exporter

sei-cosmos-exporter is a Prometheus scraper that fetches the data from a full node of a Cosmos-based blockchain via gRPC.

This is a fork of: https://github.com/solarlabsteam/cosmos-exporter 

## What can I use it for?

You can run a full node, run cosmos-exporter on the same host, set up Prometheus to scrape the data from it (see below for instructions), then set up Grafana to visualize the data coming from the exporter and probably add some alerting. Here are some examples of Grafana dashboards we created for ourselves:

![Validator dashboard](https://raw.githubusercontent.com/solarlabsteam/cosmos-exporter/master/images/dashboard_validator.png)
![Validators dashboard](https://raw.githubusercontent.com/solarlabsteam/cosmos-exporter/master/images/dashboard_validators.png)
![Wallet dashboard](https://raw.githubusercontent.com/solarlabsteam/cosmos-exporter/master/images/dashboard_wallet.png)

## How can I set it up?

First of all, you need to download the latest release from [the releases page](https://github.com/solarlabsteam/cosmos-exporter/releases/). After that, you should unzip it and you are ready to go:

```sh
bash ./install-sei-cosmos-exporter.sh
```

## How can I scrape data from it?

Here's the example of the Prometheus config you can use for scraping data:

```yaml
scrape-configs:
  # specific validator(s)
  - job_name:       'validator'
    scrape_interval: 15s
    metrics_path: /metrics/validator
    static_configs:
      - targets:
        - <list of validators you want to monitor>
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_address
      - source_labels: [__param_address]
        target_label: instance
      - target_label: __address__
        replacement: <node hostname or IP>:9300
  # specific wallet(s)
  - job_name:       'wallet'
    scrape_interval: 15s
    metrics_path: /metrics/wallet
    static_configs:
      - targets:
        - <list of wallets>
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_address
      - source_labels: [__param_address]
        target_label: instance
      - target_label: __address__
        replacement: <node hostname or IP>:9300

  # all validators
  - job_name:       'validators'
    scrape_interval: 15s
    metrics_path: /metrics/validators
    static_configs:
      - targets:
        - <node hostname or IP>:9300
```

Then restart Prometheus and you're good to go!

All of the metrics provided by cosmos-exporter have the following prefixes:
- `cosmos_validator_*` - metrics related to a single validator
- `cosmos_validators_*` - metrics related to a validator set
- `cosmos_wallet_*` - metrics related to a single wallet

## How does it work?

It queries the full node via gRPC and returns it in the format Prometheus can consume.

## How can I configure it?

You can pass the artuments to the executable file to configure it. Here is the parameters list:

- `--bech-prefix` - the global prefix for addresses. Defaults to `persistence`
- `--denom` - the currency, for example, `uatom` for Cosmos. Defaults to `uxprt`
- `--denom-coefficient` - the number of decimals, `1000000` for cosmos. Defaults to `1`. Can't provide along with `--denom-exponent`
- `--denom-exponent` - the denom exponent, `6` for cosmos. Defaults to `0`. Can't provide along with `--denom-coefficient`
- `--listen-address` - the address with port the node would listen to. For example, you can use it to redefine port or to make the exporter accessible from the outside by listening on `127.0.0.1`. Defaults to `:9300` (so it's accessible from the outside on port 9300)
- `--node` - the gRPC node URL. Defaults to `localhost:9090`
- `--tendermint-rpc` - Tendermint RPC URL to query node stats (specifically `chain-id`). Defaults to `http://localhost:26657`
- `--log-devel` - logger level. Defaults to `info`. You can set it to `debug` to make it more verbose.
- `--limit` - pagination limit for gRPC requests. Defaults to 1000.
- `--json` - output logs as JSON. Useful if you don't read it on servers but instead use logging aggregation solutions such as ELK stack.


You can also specify custom Bech32 prefixes for wallets, validators, consensus nodes, and their pubkeys by using the following params:
- `--bech-account-prefix`
- `--bech-account-pubkey-prefix`
- `--bech-validator-prefix`
- `--bech-validator-pubkey-prefix`
- `--bech-consensus-node-prefix`
- `--bech-consensus-node-pubkey-prefix`

By default, if not specified, it defaults to the next values (as it works this way for the most of the networks):
- `--bech-account-prefix` = `--bech-prefix`
- `--bech-account-pubkey-prefix` = `--bech-prefix` + "pub"
- `--bech-validator-prefix`  = `--bech-prefix` + "valoper"
- `--bech-validator-pubkey-prefix` = `--bech-prefix` + "valoperpub"
- `--bech-consensus-node-prefix` = `--bech-prefix` + "valcons"
- `--bech-consensus-node-pubkey-prefix` = `--bech-prefix` + "valconspub"

An example of the network where you have to specify all the prefixes manually is Iris, check out the flags example below.

Additionally, you can pass a `--config` flag with a path to your config file (I use `.toml`, but anything supported by [viper](https://github.com/spf13/viper) should work).

## Which networks this is guaranteed to work?

In theory, it should work on a Cosmos-based blockchains with cosmos-sdk >= 0.40.0 (that's when they added gRPC and IBC support). If this doesn't work on some chains, please file and issue and let's see what's up.

## How can I contribute?

Bug reports and feature requests are always welcome! If you want to contribute, feel free to open issues or PRs.
