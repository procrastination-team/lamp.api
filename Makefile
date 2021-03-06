CMD := api

build:
	go build -o $(CMD) ./cmd/main.go

run:
	make build
	./$(CMD)

docker:
	docker build --tag $(CMD) -f ./build/Dockerfile .

lint: 
	golangci-lint run ./...

ut:
	go test -v -count=1 -race -gcflags=-l -timeout=30s ./...

clean:
	rm $(CMD)

.PHONY: build run docker lint ut clean
