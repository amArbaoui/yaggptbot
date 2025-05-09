FROM golang:1.24.2 as builder
ENV CGO_ENABLED=1
WORKDIR /opt/yaggptbot
COPY . .
RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1
RUN make build

FROM alpine:3.20
WORKDIR /opt/yaggptbot
EXPOSE 8080
RUN apk add --no-cache tzdata
COPY --from=builder /opt/yaggptbot/app/storage storage
COPY --from=builder /opt/yaggptbot/app/db/.gitkeep db/.gitkeep
COPY --from=builder /opt/yaggptbot/main main

CMD ["./main"]
