proto_gen_all:
	protoc -I. -I=${GOPATH}/src/github.com/gogo/protobuf/protobuf -I=${GOPATH}/src \
		--gogo_out=. ScillaMessage/*.proto
	protoc -I./ZilliqaMessage -I=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
		-I=${GOPATH}/src \
		--gogo_out=ZilliqaMessage ZilliqaMessage/*.proto
