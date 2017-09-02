FROM golang:latest

WORKDIR /go/src/app
COPY main.go .

RUN go-wrapper download
RUN go-wrapper install
CMD ["go-wrapper", "run"]
