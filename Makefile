SHELL := /bin/bash

.PHONY: build
build:
	docker-compose build --no-cache

# Self-signed certs are only for testing!
.PHONY: certs
certs:
	openssl req -x509 -nodes -newkey rsa:4096 -keyout artifacts/dev-key.pem -out artifacts/dev-cert.pem -days 1000000

.PHONY: run
run: 
	docker-compose up
