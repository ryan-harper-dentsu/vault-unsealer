SHELL := /bin/bash
GO_VERSION := 1.9
VERSION := v0.1-alpha
RELEASE_NAME := ${VERSION}
RELEASE_DESCRIPTION :=
PRE_RELEASE := true

.PHONY: install
install:
	go get .
	go install

.PHONY: build-dc
build-dc:
	docker-compose build --no-cache

.PHONY: publish-release
publish-release: 
	@rm -rf release
	@go get github.com/mitchellh/gox github.com/aktau/github-release
	git tag ${VERSION} || true
	git push --tags
	@if [ "${PRE_RELEASE}" = "true" ]; then \
		pre_release_arg=--pre-release ; \
	else \
		pre_release_arg= ; \
	fi; \
	github-release release \
    --user tallpauley \
    --repo vault-unsealer \
    --tag ${VERSION} \
    --name "${RELEASE_NAME}" \
    --description "${RELEASE_DESCRIPTION}" \
	$$pre_release_arg
	@gox -output "release/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}"
	@for f in release/*; do \
		zip -j -r $$f.zip $$f; \
		github-release upload \
			--user tallpauley \
			--repo vault-unsealer \
			--tag ${VERSION} \
			--name "$$f.zip" \
			--file $$f.zip; \
	done

# Self-signed certs are only for testing!
.PHONY: certs
certs:
	openssl req -x509 -nodes -newkey rsa:4096 -keyout artifacts/dev-key.pem -out artifacts/dev-cert.pem -days 1000000

.PHONY: run-dc
run-dc: 
	docker-compose up
