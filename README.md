[![CircleCI](https://circleci.com/gh/tallpauley/vault-unsealer.svg?style=svg)](https://circleci.com/gh/tallpauley/vault-unsealer)
# Vault Unsealer

**Work in Progress! I'll take this down once it's ready to go :)**

A daemon that keeps your [Vault](https://vaultproject.io) unsealed. It's like having the required number of Vault admins always watching your Vault instance, ready to unseal it in a moment's notice.

## How it Works
- Vault admins independently and securely insert the required threshold number of unseal keys into `vault-unsealer` daemon via a **https API** (much like unsealing vault itself)
- `vault-unsealer` polls a Vault instance for [seal status](https://www.vaultproject.io/api/system/seal-status.html)
- when `vault-unsealer` detects the instance is sealed, it uses the in-memory unseal keys to unseal the vault instance

## Use Cases
- you want your single Vault instance to stay unsealed without the complications of managing an HA cluster
- you run a single Vault instance or Vault HA cluster in a preemptible environment, and want to make sure your vault instances stay unsealed through instance deletion & recreation

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
