package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	ctx := context.Background()

	h, err := createHost()
	if err != nil {
		fmt.Println("Error creating host:", err)
		return
	}
	defer h.Close()

	info, err := getServerInfo("/ip4/127.0.0.1/tcp/8080/p2p/QmeGovfdwqdYThZ5qkwAUREvnp1kkLuV6d1iMEj1MhZqcR")
	if err != nil {
		fmt.Println("Error getting server info:", err)
		return
	}

	err = connectToServer(ctx, h, info)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}

	fmt.Println("Enter:\n1 for chat stream\n2 for file share stream:")
	var choice int
	_, err = fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	switch choice {
	case 1:
		handleChatStream(ctx, h, info)
	case 2:
		handleFileShareStream(ctx, h, info)
	default:
		fmt.Println("Invalid choice")
	}
}

func createHost() (host.Host, error) {
	return libp2p.New()
}

func getServerInfo(maddrStr string) (*peer.AddrInfo, error) {
	maddr, err := ma.NewMultiaddr(maddrStr)
	if err != nil {
		return nil, err
	}

	return peer.AddrInfoFromP2pAddr(maddr)
}

func connectToServer(ctx context.Context, h host.Host, info *peer.AddrInfo) error {
	return h.Connect(ctx, *info)
}

func openStream(ctx context.Context, h host.Host, info *peer.AddrInfo, protocolID string) (network.Stream, error) {
	return h.NewStream(ctx, info.ID, protocol.ID(protocolID))
}

func readUserInput(prompt string) (string, error) {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", scanner.Err()
	}
	return scanner.Text(), nil
}

func handleChatStream(ctx context.Context, h host.Host, info *peer.AddrInfo) {
	chatStream, err := openStream(ctx, h, info, "/chat/1.0.0")
	if err != nil {
		fmt.Println("Error opening chat stream:", err)
		return
	}

	reader := bufio.NewReader(chatStream)
	writer := bufio.NewWriter(chatStream)

	for {
		message, err := readUserInput("Enter a message to send to the server: ")
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		_, err = writer.WriteString(message + "\n")
		if err != nil {
			fmt.Println("Error writing to server:", err)
			break
		}
		writer.Flush()

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
			break
		}
		fmt.Printf("Received response from server: %s\n", response)

		confirmation, err := readUserInput("Do you want to continue (yes/no)? ")
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}
		if confirmation != "yes" {
			break
		}
	}
}

func handleFileShareStream(ctx context.Context, h host.Host, info *peer.AddrInfo) {
	fileShareStream, err := openStream(ctx, h, info, "/fileshare/1.0.0")
	if err != nil {
		fmt.Println("Error opening file share stream:", err)
		return
	}

	filePath := "D:\\FTP\\repo\\LibP2P\\send.png\n"

	fileWriter := bufio.NewWriter(fileShareStream)
	_, err = fileWriter.WriteString(filePath + "\n")
	if err != nil {
		fmt.Println("Error sending file path:", err)
		return
	}
	fileWriter.Flush()

	err = os.MkdirAll("received_files", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	outFile, err := os.Create("received_files/received_file.png")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	fileReader := bufio.NewReader(fileShareStream)
	buf := make([]byte, 1024) // Size of each chunk in bytes
	for {
		n, err := fileReader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading file:", err)
			return
		}

		_, err = outFile.Write(buf[:n])
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
	}

	fmt.Println("File received successfully")
}
