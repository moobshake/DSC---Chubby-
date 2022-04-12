package nodecomm

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

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

// When passing messages between Windows and Linux, the string file paths
// use different slashes. Make sure that the file path received can be perceived
// correctly by the current machine.
// Returns nil - but changes the filepath string (ptr) to the correct one version
// the of filepath.
func getCorrectFilePath(oriFilePath *string) {
	os := runtime.GOOS
	switch os {
	case "windows":
		// fmt.Println("Windows")
		// replace all "/" to "\\" in path
		*oriFilePath = strings.Replace(*oriFilePath, "/", "\\", -1)
	default:
		// fmt.Println("Linux")
		// replace all "\\" to "/" in path
		*oriFilePath = strings.Replace(*oriFilePath, "\\", "/", -1)
	}
}
