.PHONY: test
test: lint unit build integration

.PHONY: lint
lint:
	GO111MODULE=on go vet -mod=vendor ./...

.PHONY: unit
unit:
	GO111MODULE=on go test -mod=vendor -cover -v -short ./...

.PHONY: build
build:
	mkdir -p build
	GO111MODULE=on GOOS=linux /usr/local/go/bin/go build -mod=vendor -o build/image-colors ./cmd/image-colors-lambda
	cd build && zip image-colors.zip ./image-colors
	echo "build/image-colors.zip created"

.PHONY: integration
integration:
	tests/integration/run.sh
