auth:
	go run ./cmd/auth/main.go

driver:
	go run ./cmd/driver/main.go

down:
	docker-compose down

down-v:
	docker-compose down -v

up:
	docker-compose up
