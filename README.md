## Install
```bash
cd ~ && go install github.com/bcdevtools/node-management/cmd/nmngd@latest
```

## Node setup check
- Validator node
- RPC node
- Snapshot node
- Archival node

```bash
nmngd node setup-check ~/.node_home --type validator/rpc/snapshot/archival
nmngd node extract-addrbook ~/.node_home_source/config/addrbook.json ~/.node_home_dst/config/addrbook.json
nmngd node prune-addrbook ~/.node_home/config/addrbook.json
nmngd node prune-data ~/.node_home --binary xxxd [--backup-pvs ~/priv_validator_state.json.backup]
```

## Nginx config generator

```bash
nmngd gen-nginx \
  --rpc rpc.mychain.testnet.example.com \
  --rest rest.mychain.testnet.example.com \
  --jsonrpc jsonrpc.mychain.testnet.example.com \
  [--rpc-port 26657] \
  [--rest-port 1317] \
  [--jsonrpc-port 8545]
```

## Generate SSH keys
```bash
nmngd keys add-snapshot-upload-ssh-key
# nmngd keys ss
```