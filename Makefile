SHELL := /bin/bash
GO_VERSION := 1.9

.PHONY: build
build:
	docker run --rm -v "${PWD}":/usr/src/myapp -w /usr/src/myapp golang:${GO_VERSION} go-wrapper download && go build -v

.PHONY: build-dev
build-dc:
	docker-compose build --no-cache

# Self-signed certs are only for testing!
.PHONY: certs
certs:
	openssl req -x509 -nodes -newkey rsa:4096 -keyout artifacts/dev-key.pem -out artifacts/dev-cert.pem -days 1000000

.PHONY: run-docker
run-dc: 
	docker-compose up
