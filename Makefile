.PHONY: build

build:
	GOARCH=amd64 GOOS=darwin go build -o build/client-darwin ./cmd/client
	GOARCH=amd64 GOOS=darwin go build -o build/server-darwin ./cmd/server
	GOARCH=amd64 GOOS=linux go build -o build/client-linux ./cmd/client
	GOARCH=amd64 GOOS=linux go build -o build/server-linux ./cmd/server
	GOARCH=amd64 GOOS=windows go build -o build/client-windows ./cmd/client
	GOARCH=amd64 GOOS=windows go build -o build/server-windows ./cmd/server

