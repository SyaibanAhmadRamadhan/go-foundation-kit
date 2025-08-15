up:
	docker compose up -d db
	sleep 5
	go run main.go

test:
	docker compose run --rm k6

down:
	docker compose down -v