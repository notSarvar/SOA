module promoservice/api-gateway

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/labstack/echo/v4 v4.11.4
	github.com/oapi-codegen/runtime v1.1.1
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.33.0
	promoservice/proto v0.0.0-00010101000000-000000000000
)

replace promoservice/proto => ../proto

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)
