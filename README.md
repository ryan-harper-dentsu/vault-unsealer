# Vault Unsealer

A daemon that keeps your [Vault](https://vaultproject.io) unsealed. It's like having the required number of Vault admins always watching your Vault instance, ready to unseal it in a moment's notice.

## How it Works
- Vault admins independently and securely insert the required threshold number of unseal keys into `vault-unsealer` daemon via a **https API** (much like unsealing vault itself)
- `vault-unsealer` polls a Vault instance for [seal status](https://www.vaultproject.io/api/system/seal-status.html)
- when `vault-unsealer` detects the instance is sealed, it uses the in-memory unseal keys to unseal the vault instance

## Use Cases
- you want your single Vault instance to stay unsealed without the complications of managing an HA cluster
- you run a single Vault instance or Vault HA cluster in a preemptible environment, and want to make sure your vault instances stay unsealed through instance deletion & recreation


## Installation
#### Binary
Download a binary from the [releases page](https://github.com/tallpauley/vault-unsealer/releases)
#### Docker
use the [docker image](https://hub.docker.com/r/tallpauley/vault-unsealer/)
#### From Src
Install Go 1.9 and
```make install```

## Try it out
The `docker-compose.yml` makes it easy to try `vault-unsealer`. Make sure `vault-unsealer` is installed and in your path and do the following:

1. Start vault-unsealer daemon & vault w/ `docker-compose`
```
$ make run-dc
...
vault-client_1           | Unseal Key 1: d+tlhvu4kAvIrjz5+cdYlGEbQHH694Ly+keHx4ZDlOpf
vault-client_1           | Unseal Key 2: LY/zr9n8Tu+dSgNop4NBEtm7m2kPe2fAzAlc1+HxVAmL
vault-client_1           | Unseal Key 3: PjX/3n3TkIBbT+1+6d51b3zV1y8lCk+6OyrXYuq6d9Lk
vault-unsealer_1   | 2017/09/08 21:46:28 main.go:234: 
...
vault-unsealer listening on 0.0.0.0:443
...
```
2. Switch to another terminal and set `VAULT_UNSEALER_ADDR` to enable use of `vault-unsealer` client
```
$ export VAULT_UNSEALER_ADDR=https://localhost:8443
```
3. Add the threshold of unseal keys (repeat 3 times for default Vault unseal threshold):
```
# -skip-host-verification is ONLY for testing purposes
$ vault-unsealer -add-key -skip-host-verification
Enter Unseal Key:
1 of 3 required unseal keys
...
```

4. Switch back to the terminal w/ docker-compose, and notice that the vault has been unsealed:
```
vault-unsealer_1   | 2017/09/08 21:46:50 main.go:155: Unsealing vault w/ unseal key #1
vault-unsealer_1   | 2017/09/08 21:46:50 main.go:155: Unsealing vault w/ unseal key #2
vault-unsealer_1   | 2017/09/08 21:46:50 main.go:155: Unsealing vault w/ unseal key #3
```

5. Restart the vault container and notice how vault is automatically unsealed!
```
docker-compose restart vault-server
```

## Usage
### CLI
```
$ vault-unsealer -help
Usage of vault-unsealer:
  -add-key
    	securely send an unseal key to a vault-unsealer server
  -server
    	start a vault-unsealer server
  -skip-host-verification
    	disable TLS certificate check for client commands (FOR TESTING PURPOSES ONLY)
  -status
    	view status of a vault-unsealer server
  -version
    	show version
```

### Server configuration (env variables)

`VAULT_ADDR` (required): The address of the Vault server expressed as a URL and port, for example: `http://127.0.0.1:8200`

`HTTPS_CERT` (required): path to PEM-encoded x509 certificate.

`HTTPS_CERT_KEY` (required): path to PEM-encoded private key

`LISTEN_ADDR` (optional): String specifying host and port to listen on. Default is `:443`. See https://golang.org/pkg/net/#Listen

`POLLING_INTERVAL` (optional): Amount of seconds between checking on Vault's `seal-status`. Default is `1`

`ROOT_TOKEN_TEST` (optional): Optionally, a root token can be generated and destroyed immediately after to check if the unseal keys are valid. Values are `true` or `false`. Default is `false`.

### Client configuration (env variables)

`VAULT_UNSEALER_ADDR`: Hostname and port of vault-unsealer server/daemon. Such as `https://localhost:8443`.


## Tradeoffs
You have to be willing to accept:
- extra/alternate procedures to seal the vault (since vault-unsealer constantly unseals) such as:
  - shutting down `vault-unsealer` daemon(s) before sealing Vault
  - OR: we cut off access to Vault using an alternate method (like shutting down the Vault instance)
- `vault-unsealer` polls on an interval and unseals a Vault instance accordingly-- you can't expect Vault to be unsealed 100% of the time, or immediate failover like when using an Vault HA storage backend
- `vault-unsealer` itself loses it's unseal keys (in memory) when it restarts (like Vault), so ideally `vault-unsealer` runs somewhere where restarts are less frequent

## Security
* Unseal keys are only in memory, inserted via `https` API
* Like Vault, it attempts to use `mlock` syscall to prevent unseal keys from being swapped to disk
