CMD := api

run:
	docker-compose -f ./deployment/docker-compose.yaml up

docker:
	docker build --tag $(CMD) -f ./Dockerfile .
	docker run -p 80:80 $(CMD)

build:
	go build -o $(CMD) ./cmd/main.go

lint: 
	golangci-lint run ./...

ut:
	go test -v -count=1 -race -gcflags=-l -timeout=30s ./...

clean:
	rm $(CMD)

.PHONY: build run docker lint ut clean
