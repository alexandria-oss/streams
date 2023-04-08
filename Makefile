unit-test:
	go test ./... -coverprofile coverage.out .

unit-test-html: unit-test
	go tool cover -html=coverage.out

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

gen-proto: $(GO_BUILD_DIR)/$(TARGET_EXEC)

publish-pkg:
	./publish-go-pkg.sh -v $(version) -m $(module_name)
