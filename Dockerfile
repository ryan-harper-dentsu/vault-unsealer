# Multi-stage build (requires Docker 17.05 or greater)
# for first stage of canary container, we just build the binary
FROM golang:1.9

WORKDIR /go/src/app
ADD main.go .

RUN go-wrapper download
RUN CGO_ENABLED=0 go build -v 

# second stage is just copy the binary from the first stage
FROM alpine:3.6
WORKDIR /

RUN apk update && apk add --no-cache ca-certificates

COPY --from=0 /go/src/app/app /vault-unsealer
RUN chmod +x /vault-unsealer

EXPOSE 443

ENTRYPOINT ["/vault-unsealer"]
CMD ["-server"]