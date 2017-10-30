# Comparison to other Vault "unsealers"
Here is a really basic comparison of my vault-unsealer to some other vault unsealers out there. Let me know if these need to be updated!

The primary advantage of this unsealer is that unseal keys are stored in memory. This makes it hard for even other vault admins to read your unseal key.

| tool                                    | unseal key storage model      | unseal multiple instances | limitations                                                       |
|-----------------------------------------|:-----------------------------:|--------------------------:|-------------------------------------------------------------------|
| [tallpauley/vault-unsealer]             | in-memory                     | planned v0.2: only HA     | manual add-key process, unseal keys don't persist across restarts |
| [jaxxstorm/hookpick]                    | PGP-encrypted in config file  | yes                       | manual PGP passcode entry, PGP private key(s) on server           | 
| [jetstack-experimental/vault-unsealer]  | Amazon/Google KMS/GCS Bucket  | no                        | unseal keys visible to those with KMS/bucket access               |
| [InQuicker/vault-auto-unsealer]         | environment variable          | no                        | only supports single unseal key                                   |
| [blockloop/vault-unseal]                | environment variables         | no                        | unseal keys visible to server admin                               | 

[tallpauley/vault-unsealer]: https://github.com/tallpauley/vault-unsealer
[jetstack-experimental/vault-unsealer]: https://github.com/jetstack-experimental/vault-unsealer
[InQuicker/vault-auto-unsealer]: https://github.com/InQuicker/vault-auto-unsealer
[blockloop/vault-unseal]: https://hub.docker.com/r/blockloop/vault-unseal/
[jaxxstorm/hookpick]: https://github.com/jaxxstorm/hookpick