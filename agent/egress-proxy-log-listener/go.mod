module github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener

go 1.19

require (
	github.com/alexandria-oss/streams v0.0.0-20230308031338-8e2e9d5dcb07
	github.com/alexandria-oss/streams/driver/kafka v0.0.0-20230320031154-f7c183d65d17
	github.com/alexandria-oss/streams/driver/sql v0.0.0-00010101000000-000000000000
	github.com/jackc/pglogrepl v0.0.0-20230318140337-5ef673a9d169
	github.com/jackc/pgx/v5 v5.3.1
	github.com/rs/zerolog v1.29.0
	github.com/segmentio/kafka-go v0.4.39
	github.com/spf13/viper v1.15.0
)

require (
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/protobuf v1.29.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/alexandria-oss/streams => ../../

replace github.com/alexandria-oss/streams/driver/sql => ../../driver/sql

replace github.com/alexandria-oss/streams/driver/kafka => ../../driver/kafka
