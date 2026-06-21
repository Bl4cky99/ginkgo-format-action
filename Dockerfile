FROM golang:1.26 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /ginkgo-format-action .

FROM gcr.io/distroless/base-debian12
COPY --from=builder /ginkgo-format-action /usr/local/bin/ginkgo-format-action

ENTRYPOINT ["/usr/local/bin/ginkgo-format-action"]
