module editor

go 1.27.1

require (
	github.com/distributed-programming-2026/go-sdk/pkg/amqp v0.0.3
	github.com/distributed-programming-2026/go-sdk/pkg/event v0.0.3
	github.com/distributed-programming-2026/go-sdk/pkg/logging v0.0.3
	github.com/distributed-programming-2026/go-sdk/pkg/migrator v0.0.3
	github.com/distributed-programming-2026/go-sdk/pkg/mysql v0.0.5
	github.com/distributed-programming-2026/go-sdk/pkg/uow v0.0.3
	github.com/distributed-programming-2026/lib v0.0.0
	github.com/getkin/kin-openapi v0.149.0
	github.com/go-chi/chi/v5 v5.3.2
	github.com/go-sql-driver/mysql v1.10.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/oapi-codegen/runtime v1.7.0
	github.com/stretchr/testify v1.12.1
	github.com/urfave/cli/v3 v3.14.0
	golang.org/x/sync v0.22.0
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
	user v0.0.0
)

replace user => ../user

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/distributed-programming-2026/go-sdk/internal/sharedpool v0.0.5 // indirect
	github.com/go-openapi/jsonpointer v0.22.5 // indirect
	github.com/go-openapi/swag/jsonname v0.25.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/oasdiff/yaml v0.1.1 // indirect
	github.com/oasdiff/yaml3 v0.0.14 // indirect
	github.com/rabbitmq/amqp091-go v1.15.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260706201446-f0a921348800 // indirect
)

replace github.com/distributed-programming-2026/lib => ../../lib
