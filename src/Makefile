BIN=template-go

build: lint
	@go mod tidy -v
	go build -o ../../bin/${BIN} main.go

run: build
	../../bin/${BIN}

lint:
	golangci-lint run

clean:
	rm ../../bin/${BIN}
