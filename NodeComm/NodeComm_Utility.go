package nodecomm

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//connectTo abstracts three lines of grpc client code
func connectTo(address, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(address+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return conn, err
}

// Returns the checksum of a given file
func getFileChecksum(fullFilePath string) []byte {
	fileToCheck, err := os.Open(fullFilePath)
	if err != nil {
		fmt.Println("getFileChecksum ERROR:", err)
		return []byte{}
	}
	defer fileToCheck.Close()

	checksum := sha256.New()
	if _, err := io.Copy(checksum, fileToCheck); err != nil {
		fmt.Println("getFileChecksum ERROR:", err)
		return []byte{}
	}
	return checksum.Sum(nil)
}

func checkFileSumSame(chksum1, chksum2 []byte) bool {
	if len(chksum1) != len(chksum2) {
		return false
	}
	for i := 0; i < len(chksum1); i++ {
		if chksum1[i] != chksum2[i] {
			return false
		}
	}
	return true
}
