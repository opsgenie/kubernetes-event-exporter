FROM golang:1.14 AS builder

ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GO11MODULE=on go build -mod=vendor -a -o /main .

FROM gcr.io/distroless/base
COPY --from=builder /main /kubernetes-event-exporter
ENTRYPOINT ["/kubernetes-event-exporter"]
