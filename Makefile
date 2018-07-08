all:
	protoc -I mr/ mr/mr.proto --go_out=plugins=grpc:mr
	protoc -I ms/ ms/ms.proto --go_out=plugins=grpc:ms
	protoc -I stldice/ stldice/stldice.proto --go_out=plugins=grpc:stldice
	protoc -I stl2svx/proto/ stl2svx/proto/stl2svx.proto --go_out=plugins=grpc:stl2svx/proto
