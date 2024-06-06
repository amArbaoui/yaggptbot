FROM golang:1.22.4 as builder
ENV CGO_ENABLED=1
WORKDIR /opt/yaggptbot
COPY . .
RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1
RUN make migrate-up
RUN make build

FROM alpine:3.20
WORKDIR /opt/yaggptbot
EXPOSE 8080
COPY --from=builder /opt/yaggptbot/main main
COPY --from=builder /opt/yaggptbot/app/.env .env
COPY --from=builder /opt/yaggptbot/app/db/yaggptbot.db ./db/yaggptbot.db

CMD ["./main"]
