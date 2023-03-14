unit-test:
	go test ./... -coverprofile coverage.out .

unit-test-html: unit-test
	go tool cover -html=coverage.out

integration-test:
	go test ./... --tags=integration -coverprofile coverage.out .

integration-test-html: integration-test
	go tool cover -html=coverage.out