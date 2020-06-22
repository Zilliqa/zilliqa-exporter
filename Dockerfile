FROM golang:1.14.4-alpine AS builder
WORKDIR /exporter
COPY . .

RUN apk add --no-cache make git; make local

FROM nginx
COPY --from=builder /exporter/dist/genet-exporter /
