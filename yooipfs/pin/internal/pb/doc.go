package pb

//go:generate protoc --gogo_out=. header.proto

// kludge to get vendoring right in protobuf output
//go:generate sed -i s,github.com/,github.com/yooba-team/yooba/yooipfs/Godeps/_workspace/src/github.com/,g header.pb.go
