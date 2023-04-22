module sample

go 1.18

require (
	github.com/alexandria-oss/streams v0.0.1-alpha.5
	github.com/alexandria-oss/streams/driver/kafka v0.0.0-20230320031154-f7c183d65d17
	github.com/alexandria-oss/streams/driver/sql v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.3.0
	github.com/json-iterator/go v1.1.12
	github.com/lib/pq v1.10.7
	github.com/modern-go/reflect2 v1.0.2
	github.com/segmentio/kafka-go v0.4.39
)

require (
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	google.golang.org/protobuf v1.29.0 // indirect
)

replace github.com/alexandria-oss/streams => ../../

replace github.com/alexandria-oss/streams/driver/sql => ../../driver/sql

replace github.com/alexandria-oss/streams/driver/kafka => ../../driver/kafka
