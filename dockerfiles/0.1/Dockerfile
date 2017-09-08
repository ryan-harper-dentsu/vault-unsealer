# Multi-stage build (requires Docker 17.05 or greater)
# first stage of build is to fetch the zip and unzip it
FROM alpine:3.6
MAINTAINER Chris Pauley <cmpauley@stanford.edu> (@tallpauley)

ENV VAULT_UNSEALER_VERSION v0.1-alpha

WORKDIR /
RUN apk update && apk add unzip
RUN wget https://github.com/tallpauley/vault-unsealer/releases/download/${VAULT_UNSEALER_VERSION}/release.vault-unsealer_linux_amd64.zip
RUN unzip release.vault-unsealer_linux_amd64.zip

# second stage is just copy the binary from the first stage
FROM alpine:3.6
WORKDIR /
COPY --from=0 /vault-unsealer .

EXPOSE 443

ENTRYPOINT ["vault-unsealer"]
CMD ["-server"]
