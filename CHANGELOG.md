# v0.2 (unreleased)
* Compare to other vault unsealers in documentation (#1)
* Add environment variables to choose TLS versions and allowed cipher suites (#2)

# v0.1 (unreleased)
* fixed bug in which vault-unsealer would become unresponsive due to polling loop
* fixed bug in which server mode wouldn't complain if memory couldn't be locked
* main.go sha1sum is passed as main.Version parameter, allowing for easier identification of canary builds
* added check for duplicate unseal keys
* adding code documentation

# v0.1-alpha
* initial prototype
