FROM golang:1.16.6-alpine3.13 AS builder

ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO11MODULE=on go build -mod=vendor -a -o /main .

FROM 192.168.1.52/system_containers/busybox:latest
COPY --from=builder /main /kubernetes-event-exporter
ENTRYPOINT ["/kubernetes-event-exporter"]
