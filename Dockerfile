FROM golang:1.9

WORKDIR /go/src/app
COPY main.go .

RUN go-wrapper download
RUN go-wrapper install
CMD ["go-wrapper", "run"]
