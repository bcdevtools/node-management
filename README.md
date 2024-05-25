## Install
```bash
git clone https://github.com/bcdevtools/node-management
cd node-management
make install
```

## Node setup check
- Validator node
- RPC node
- Snapshot node
- Archival node

```bash
nmngd node setup-check ~/.node_home --type validator/rpc/snapshot/archival
```

## Node management

```bash
nmngd node extract-addrbook ~/.node_home_source/config/addrbook.json ~/.node_home_dst/config/addrbook.json
nmngd node prune-addrbook ~/.node_home/config/addrbook.json
nmngd node prune-data ~/.node_home --binary xxxd [--backup-pvs ~/priv_validator_state.json.backup]
nmngd node state-sync ~/.node_home --binary xxxd --rpc http://localhost:26657 [--address-book /home/x/.node/config/addrbook.json] [--peers nodeid@127.0.0.1:26656] [--seeds seed@1.1.1.1:26656] [--max-duration 12h]
nmngd node dump-snapshot ~/.node_home --binary xxxd [--max-duration 12h] [--no-service] [--service-name xxx] [--external-rpc https://rpc1.example.com:443 --external-rpc https://rpc2.example.com:443]
nmngd node zip-snapshot ~/.node_home
```

### For validator node
```bash
nmngd node auto-backup-priv-validator-state-json ~/.node_home --binary xxxd
# generate setup for auto-backup-pvs
nmngd node auto-backup-priv-validator-state-json ~/.node_home --binary xxxd --gen-setup
```

## Run web server
```bash
nmngd start-web ~/.rpc-gaia \
  --port 8080 \
  --authorization-token "X" \
  --chain-name "Cosmos Hub" \
  --chain-description "Multi-lines describes the chain\nand its features" \
  --chain-id cosmoshub-4 \
  --g-binary-name gaid \
  --g-node-home-name .gaia \
  --snapshot-file /snapshot/cosmos-mainnet/snapshot.tar.lz \
  --snapshot-download-url https://cosmos.m.valoper.io/snapshot/snapshot.tar.lz \
  --exr-rpc-url https://rpc1.cosmos.m.valoper.io \
  --exr-rest-url https://rest1.cosmos.m.valoper.io \
  --exr-favicon-url https://cosmos.m.valoper.io/favicon.ico \
  --exr-logo-url https://cosmos.m.valoper.io/logo.png \
  --monitor-disks /mount/data1 --monitor-disks /mount/data2
```
Generate start command:
```bash
nmngd gen-start-web
```

## Nginx config generator

```bash
nmngd gen-nginx \
  --rpc rpc.mychain.testnet.example.com \
  --rest rest.mychain.testnet.example.com \
  --jsonrpc jsonrpc.mychain.testnet.example.com \
  --web mychain.testnet.example.com \
  [--rpc-port 26657] \
  [--rest-port 1317] \
  [--jsonrpc-port 8545] \
  [--web-port 8080]
```

## Generate SSH keys
```bash
nmngd keys add-snapshot-upload-ssh-key
# nmngd keys ss
```