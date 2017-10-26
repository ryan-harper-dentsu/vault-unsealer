# v0.2 (unreleased)
* Add environment variables to choose TLS versions and allowed cipher suites (#2)

# v0.1 (unreleased)
* bugfix: vault-unsealer would become unresponsive due to polling loop
* bugfix: server mode wouldn't complain if memory couldn't be locked
* bugfix: no newline after client prints (#3)
* main.go sha1sum is passed as main.Version parameter, allowing for easier identification of canary builds
* added check for duplicate unseal keys
* adding code documentation
* Compare to other vault unsealers in documentation (#1)

# v0.1-alpha
* initial prototype
