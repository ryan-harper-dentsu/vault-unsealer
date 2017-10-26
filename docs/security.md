# Security

## Security Model

* HTTPS is required for encryption and host verification (host verification can be disabled by client for local testing)
* The API only allows for unseal key **insertion**, not **extraction**.
* Root token generation is optional, known only to vault-unsealer, and thrown away immediately after being generated
    * Also, the only time a token is used for communication with vault is root token revocation.
* Unseal keys or root tokens are **never** logged.
* Like Vault, it attempts to use `mlock` syscall to prevent unseal keys from being swapped to disk

## API Security

The vault-unsealer API is **unauthenticated**, since it only provides endpoints to **add** an unseal key, and **extract** some trivial information.

The only information that the following endpoints provide is:
* `/add-key`
    * whether a given unseal key has been already been added to **vault-unsealer** (only until threshold is met)
    * how many unseal keys have been added to **vault-unsealer** (only until threshold is met)
    * how many unseal keys vault requires (public information from vault)
* `/status`
    * how many unseal keys have been added to vault-unsealer
    * how many unseal keys vault requires (public information from vault)

## Security Best Practices

* the vault server itself should have HTTPS enabled
* `-skip-host-verification` option should only ever be used for local testing

## Security FAQ

* Why is the API unauthenticated?
  * see **API Security** section
* Couldn't an attacker use vault-unsealer server to try out unseal keys against the vault server?
  * yes, but the attacker could just directly try unseal keys against vault itself. Unsealing and root token generation APIs in Vault aren't authenticated with tokens.
