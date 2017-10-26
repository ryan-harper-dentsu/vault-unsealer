# Comparison to other vault "unsealers"
Here is a really basic comparison of my vault-unsealer to some other vault "unsealers" out there. Let me know if these need to be updated!

|  tool  |  unseal key storage   |  limitations |
|-----------------------------------------|:-----------------------------:|--------------------------------:|
| [tallpauley/vault-unsealer]             | in-memory                     | unseal keys don't persist across restarts | 
| [jetstack-experimental/vault-unsealer]  | Amazon/Google KMS/GCS Bucket  | 
| [InQuicker/vault-auto-unsealer]         | environment variable          |  only supports single unseal key |
| [blockloop/vault-unseal]                | environment variables         |                                  | 

[tallpauley/vault-unsealer]: https://github.com/tallpauley/vault-unsealer
[jetstack-experimental/vault-unsealer]: https://github.com/jetstack-experimental/vault-unsealer
[InQuicker/vault-auto-unsealer]: https://github.com/InQuicker/vault-auto-unsealer
[blockloop/vault-unseal]: https://hub.docker.com/r/blockloop/vault-unseal/