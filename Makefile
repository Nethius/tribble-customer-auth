BINARY_NAME=trimble-auth-server

build:
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME} cmd/main.go
	cp .env bin/.env

run:
	./bin/${BINARY_NAME}

build_and_run: build run

clean:
	go clean
	rm bin/${BINARY_NAME}
	rm bin/.env
