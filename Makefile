build-server:
	@go build -o bin/CacheD ./cmd/CacheD

build-client:
	@go build -o bin/client ./cmd/client

build:
	@go build -o bin/CacheD ./cmd/CacheD
	@go build -o bin/client ./cmd/client

clean:
	@rm -rf bin/