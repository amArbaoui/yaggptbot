    FROM golang:1.24.2 as builder

    ENV CGO_ENABLED=1
    ENV APP_DIR=/opt/yaggptbot

    WORKDIR ${APP_DIR}

    COPY . .

    RUN make build
    RUN ls -lR

    FROM alpine:3.20

    ENV APP_DIR=/opt/yaggptbot
    
    RUN addgroup -S appgroup && adduser -S -u 1001 -G appgroup appuser && \
        mkdir -p ${APP_DIR}/storage ${APP_DIR}/db && \
        chown -R appuser:appgroup ${APP_DIR}

    RUN apk add --no-cache tzdata

    WORKDIR ${APP_DIR}

    COPY --from=builder --chown=appuser:appgroup ${APP_DIR}/app/db/.gitkeep db/.gitkeep
    COPY --from=builder --chown=appuser:appgroup ${APP_DIR}/app/storage storage
    COPY --from=builder --chown=appuser:appgroup ${APP_DIR}/main .

    EXPOSE 8080

    USER appuser

    CMD ["./main"]
