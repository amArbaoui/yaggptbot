migrate-up:
	migrate -path=app/storage/migrations -database "sqlite3://app/db/yaggptbot.db" -verbose up
migrate-down:
	migrate -path=app/storage/migrations -database "sqlite3://app/db/yaggptbot.db" -verbose down

migrate-fix:
	CGO_ENABLED=1 migrate -path=app/storage/migrations -database "sqlite3://app/db/yaggptbot.db" -force VERSION

build:
	CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main  -ldflags='-w -s -extldflags "-static"' app/main.go