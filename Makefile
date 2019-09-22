up:
	docker-compose up -d --build

down:
	docker-compose down

test:
	docker exec chirper go test ./...
