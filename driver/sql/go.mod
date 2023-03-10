module github.com/alexandria-oss/streams/driver/sql

go 1.18

replace github.com/alexandria-oss/streams => ../../

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/alexandria-oss/streams v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.3.0
	github.com/lib/pq v1.10.7
	github.com/segmentio/ksuid v1.0.4
	github.com/stretchr/testify v1.8.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	google.golang.org/protobuf v1.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
