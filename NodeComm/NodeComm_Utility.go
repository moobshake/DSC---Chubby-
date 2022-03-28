package nodecomm

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//connectTo abstracts three lines of grpc client code
func connectTo(address, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(address+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return conn, err
}
