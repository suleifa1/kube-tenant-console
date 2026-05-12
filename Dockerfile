FROM golang:1.26.2-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/kube-tenant-console ./cmd/server

FROM alpine:3.20
RUN adduser -D -H -u 10001 app
USER 10001
WORKDIR /app
COPY --from=build /out/kube-tenant-console /app/kube-tenant-console
EXPOSE 8080
ENTRYPOINT ["/app/kube-tenant-console"]
