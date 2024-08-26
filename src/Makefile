BIN_DIR = bin
BIN = template-go
TARGET = ../${BIN_DIR}/${BIN}
SRC = main.go
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
