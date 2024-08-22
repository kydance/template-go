BIN_DIR = bin
SRC = main.go
TARGET = ../${BIN_DIR}/template-go

GO = go build -o

build: lint
	@go mod tidy -v
	${GO} ${TARGET} ${SRC}

run: build
	${TARGET}

lint:
	golangci-lint run

clean:
	rm -f ${TARGET}
