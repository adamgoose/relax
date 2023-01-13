FROM golang:1.19-alpine3.17 as builder
RUN apk add --update --no-cache alpine-sdk

WORKDIR /work
COPY go.mod go.sum /work/
RUN go mod download

COPY . /work
RUN CGO_ENABLED=0 go build -ldflags="-s -w" && \
  strip ./relax

FROM alpine:3.17
RUN apk add --update --no-cache bash ca-certificates

COPY --from=builder /work/relax /usr/local/bin/relax
CMD ["relax"]
