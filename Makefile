migrateup:
	migrate -path db/migration -database "postgresql://neondb_owner:npg_AeWjEvOz65HK@ep-empty-bar-a11p3ep6-pooler.ap-southeast-1.aws.neon.tech/urlshortsevice?sslmode=require&channel_binding=require" -verbose up

createschema:
	migrate create -ext sql -dir db/migration -seq init_schema

sqlc:
	sqlc generate

PHONY: createschema migrateup sqlc