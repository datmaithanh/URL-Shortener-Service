include .env.prod
export DB_SOURCE

migrateup:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose up

createschema:
	migrate create -ext sql -dir db/migration -seq init_schema

sqlc:
	sqlc generate


run: 
	go run cmd/main/main.go

PHONY: createschema migrateup sqlc run