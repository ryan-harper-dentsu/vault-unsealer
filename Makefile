SHELL := /bin/bash
GO_VERSION := 1.9
VERSION := v0.1
RELEASE_NAME := ${VERSION}
RELEASE_DESCRIPTION :=
PRE_RELEASE := false

.PHONY: install
install:
	go get .
	go install

.PHONY: build-dc
build-dc:
	docker-compose build --no-cache

.PHONY: build-release
build-release:
	@rm -rf release
	@gox -output "release/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}" -ldflags "-X main.Version=${VERSION}"

.PHONY: publish-release
publish-release: build-release
	@if [ -z "${GITUB_TOKEN}" ]; then >&2 echo "ERROR: GITHUB_TOKEN not set"; exit 1; fi
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
