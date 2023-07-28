I = "⚪"
E = "🔴"
D = "🔵"

default:
	@echo "$(D) supported commands: [init, update, test]"

init:
	@echo "$(I) initialiazing..."
	@rm -rf go.mod go.sum ./vendor ./mocks
	@go mod init $$(pwd | awk -F'/' '{print $$NF}') || (echo "$(E) initialization error"; exit 1)

update:
	@echo "$(I) installing dependencies..."
	@go get ./... || (echo "$(E) 'go get' error"; exit 1)
	@go get -u github.com/stretchr/objx || (echo "$(E) 'go get' error"; exit 1)
	@echo "$(I) updating imports..."
	@go mod tidy || (echo "$(E) 'go mod tidy' error"; exit 1)
	@echo "$(I) vendoring..."
	@go mod vendor || (echo "$(E) 'go mod vendor' error"; exit 1)
	@echo "$(I) regenerating mocks package..."
	@mockery --name=Client --dir=repository/api/
	# @mockery --name=<interface-name> --dir=vendor/github.com/<org>/<proj>/

test:
	@echo "$(I) linting..."
	@golangci-lint run --skip-dirs=mocks --skip-dirs=vendor || (echo "$(E) linter error"; exit 1)
	@echo "$(I) unit testing (approximately 1 minute)..."
	@go test -v $$(go list ./... | grep -v vendor | grep -v mocks) -race -coverprofile=coverage.txt -covermode=atomic

codecov: test
	@echo "$(I) analyzing coverage..."
	@go tool cover -html=coverage.txt || (echo "$(E) 'go tool cover' error"; exit 1)