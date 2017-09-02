[![CircleCI](https://circleci.com/gh/tallpauley/vault-unsealer.svg?style=svg)](https://circleci.com/gh/tallpauley/vault-unsealer)
# Vault Unsealer

Work in Progress! I'll take this down once it's ready to go :)

A daemon that keeps your [vault](https://vaultproject.io) unsealed

## How it Works
- users independently and securely insert the threshold of unseal keys into `vault-unsealer` daemon via a **https api** (much like unsealing vault itself)
- `vault-unsealer` polls a vault instance for [seal status](https://www.vaultproject.io/api/system/seal-status.html)
- when `vault-unsealer` detects the instance is sealed, it uses the in-memory unseal keys to unseal the vault instance

## Motivations
- We want a vault instance (or cluster) to *always be unsealed*, even if *every instance* is restarted
- We want an single vault instance to be relatively highly-available, without work of setting up an HA backend such as **consul** or **etcd**
  - though `vault-unsealer` can be used for an HA vault setup too!

## Tradeoffs
You have to be willing to accept:
- extra/alternate procedures to seal the vault such as:
  - shutting down `vault-unsealer` instance(s) before sealing vault
  - OR: we stop access to vault using an alternate method like destroying the instance(s) temporarily (works in a codified environment like Kubernetes)
- vault unseal keys being stored in memory in `vault-unseal` instances

## Security
* Unseal keys are only in memory, transferred by secure `https`
* Like **vault**, it attempts to use `mlock` syscall to prevent unseal keys from being swapped to disk

## Comparison to running Vault w/ HA backend
Note: a primary motivating factor of writing **vault-unsealer** is running on Kubernetes in which it is normal for instances to die for a number of reasons (cluster upgrade, etc). This significantly influences the below comparison.

| Feature                                          | Vault HA Cluster              | Single Vault w/ vault-unsealer
| ------------------------------------------------ |:-----------------------------:| :-------------------------:|
| behavior if ALL vault instances die              | no unsealed vault             | unsealed vault available shortly
| failover time                                    | instant                | vault restart + Poll Interval + Unseal time

If you want instant failover, and always unsealed vault instances, you can combine `vault-unsealer` with an HA backend

## The Big Question
**But `vault-unsealer` itself can be restarted too! Doesn't HA solve this?**

Yes, but it is more lightweight & easier to distribute `vault-unsealer` daemons across datacenters than it is to maintain a resilient HA cluster backend. It can simply increase the unsealed uptime of a single vault instance or cluster.