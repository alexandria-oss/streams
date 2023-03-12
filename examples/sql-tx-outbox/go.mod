module sample

go 1.18

require (
	github.com/alexandria-oss/streams v0.0.0-20230308031338-8e2e9d5dcb07
	github.com/alexandria-oss/streams/driver/sql v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.3.0
	github.com/json-iterator/go v1.1.12
	github.com/lib/pq v1.10.7
)

require (
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	google.golang.org/protobuf v1.29.0 // indirect
)

replace github.com/alexandria-oss/streams => ../../

replace github.com/alexandria-oss/streams/driver/sql => ../../driver/sql
