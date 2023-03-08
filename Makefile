unit-test-cov:
	go test ./... -coverprofile coverage.out .

unit-test-cov-html: unit-test-cov
	go tool cover -html=coverage.out