module github.com/qgxpagamentos/nrfasthttp

go 1.18

require (
	github.com/newrelic/go-agent/v3 v3.16.1
	github.com/valyala/fasthttp v1.37.0
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.39.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

retract (
	// Mistake happened in the version DO NOT USE
	v1.0.0
	// Mistake happened in the version DO NOT USE
	v1.0.1
	// Mistake happened in the version DO NOT USE
	v1.0.2
)
