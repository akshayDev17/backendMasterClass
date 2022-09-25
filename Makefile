postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=mysecretpassword -d postgres:latest
createdb:
	docker exec -it postgres createdb --username postgres --owner=postgres simple_bank
connect_to_db:
	docker exec -it postgres psql -U postgres -d simple_bank
dropdb:
	docker exec -it postgres dropdb simple_bank
migrate_up:
	migrate -path db/migration -database "postgresql://postgres:mysecretpassword@localhost:5432/simple_bank?sslmode=disable" -verbose up
migrate_down:
	migrate -path db/migration -database "postgresql://postgres:mysecretpassword@localhost:5432/simple_bank?sslmode=disable" -verbose down
migrate_force_v1:
	migrate -path db/migration -database "postgresql://postgres:mysecretpassword@localhost:5432/simple_bank?sslmode=disable" -verbose force 1
sqlc: 
	sqlc generate
test:
	go test -v -cover ./...
.PHONY: postgres createdb dropdb migrate_up sqlc