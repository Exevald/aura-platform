package main

import (
	filestorageapi "filestorage/api/server/internalapi"
	"publication/internal/publication/app/adapter"
	fileclient "publication/internal/publication/infra/adapter"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func connectFiles(address string) (adapter.FilestorageService, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return fileclient.NewFilestorageService(filestorageapi.NewFilestorageServiceClient(conn)), conn, nil
}
