TARGET_EXEC := alexandria

GO_BUILD_DIR := ./persistence
SRC_DIR := ./

# Find all the C and C++ files we want to compile
# Note the single quotes around the * expressions. Make will incorrectly expand these otherwise.
SRCS := $(shell find $(SRC_DIR) -name '*.proto')

$(GO_BUILD_DIR)/$(TARGET_EXEC): $(SRCS)
	protoc -I. --go_out=$(GO_BUILD_DIR) $(SRCS)
	rsync -a $(GO_BUILD_DIR)/github.com/alexandria-oss/streams/persistence/ $(GO_BUILD_DIR)
	rm -r $(GO_BUILD_DIR)/github.com

unit-test-cov:
	go test ./... -coverprofile coverage.out .

unit-test-cov-html: unit-test-cov
	go tool cover -html=coverage.out