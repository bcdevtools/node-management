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
nmngd check ~/.node_home --type validator/rpc/snapshot/archival
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
