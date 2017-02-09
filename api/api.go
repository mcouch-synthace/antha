//go:generate protoc -I/usr/local/include -I${GOPATH}/src -I. ${GOPATH}/src/github.com/antha-lang/antha/api/v1/coord.proto ${GOPATH}/src/github.com/antha-lang/antha/api/v1/inventory.proto ${GOPATH}/src/github.com/antha-lang/antha/api/v1/measurement.proto ${GOPATH}/src/github.com/antha-lang/antha/api/v1/message.proto ${GOPATH}/src/github.com/antha-lang/antha/api/v1/state.proto ${GOPATH}/src/github.com/antha-lang/antha/api/v1/task.proto --go_out=plugins=grpc:${GOPATH}/src
package api
