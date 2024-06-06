migrate-up:
	migrate -path=app/storage/migrations -database "sqlite3://app/yaggptbot.db" -verbose up
migrate-down:
	migrate -path=app/storage/migrations -database "sqlite3://app/yaggptbot.db" -verbose down

migrate-fix:
	migrate -path=app/storage/migrations -database "sqlite3://app/yaggptbot.db" -force VERSION
