BINARY_NAME=deadlinks
BINARY_PATH=./bin/$(BINARY_NAME)

build:
	@go build -o $(BINARY_PATH) main.go 

run: build
	@$(BINARY_PATH) --url $(url)

clean:
	@rm -rf ./bin 

dev:
	@go run main.go --url $(url)
