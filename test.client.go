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

	h := createHost()
	defer h.Close()

	info := getServerInfo("/ip4/127.0.0.1/tcp/8080/p2p/QmTACRYYczNc7qfxpseR1Meac6NHiSbCczvjbJvMHGWjsA")

	connectToServer(ctx, h, info)
	// if err != nil {
	// 	fmt.Println("Error connecting to server:", err)
	// 	return
	// }

	fmt.Println("Enter:\n1 for chat stream\n2 for file share stream:")
	var choice int
	_, err := fmt.Scan(&choice)
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

func createHost() host.Host {
	h, err := libp2p.New()
	if err != nil {
		panic(err)
	}
	return h
}

func getServerInfo(maddrStr string) *peer.AddrInfo {
	maddr, err := ma.NewMultiaddr(maddrStr)
	if err != nil {
		panic(err)
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		panic(err)
	}

	return info
}

func connectToServer(ctx context.Context, h host.Host, info *peer.AddrInfo) {
	err := h.Connect(ctx, *info)
	if err != nil {
		panic(err)
	}
}

func openStream(ctx context.Context, h host.Host, info *peer.AddrInfo, protocolID string) network.Stream {
	s, err := h.NewStream(ctx, info.ID, protocol.ID(protocolID))
	if err != nil {
		panic(err)
	}
	return s
}

func handleChatStream(ctx context.Context, h host.Host, info *peer.AddrInfo) {
	chatStream := openStream(ctx, h, info, "/chat/1.0.0")
	reader := bufio.NewReader(chatStream) // Read from standard input
	writer := bufio.NewWriter(chatStream)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter a message to send to the server: ")
		if !scanner.Scan() {
			fmt.Println("Error reading input:", scanner.Err())
			break
		}
		message := scanner.Text()

		// Send the message to the server
		var err error
		_, err = writer.WriteString(message + "\n")
		if err != nil {
			fmt.Println("Error writing to server:", err)
			break
		}
		writer.Flush()

		// Read the server's response
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
			break
		}
		fmt.Printf("Received response from server: %s\n", response)

		fmt.Print("Do you want to continue (yes/no)? ")
		if !scanner.Scan() {
			fmt.Println("Error reading input:", scanner.Err())
			break
		}
		confirmation := scanner.Text()
		if confirmation != "yes" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}

func handleFileShareStream(ctx context.Context, h host.Host, info *peer.AddrInfo) {

	fileShareStream := openStream(ctx, h, info, "/fileshare/1.0.0")
	// reader := bufio.NewReader(chatStream)
	// writer := bufio.NewWriter(chatStream)

	fileWriter := bufio.NewWriter(fileShareStream)
	_, err := fileWriter.WriteString("D:\\FTP\\repo\\LibP2P\\send.png\n")
	if err != nil {
		fmt.Println("Error sending file path:", err)
		return
	}
	fileWriter.Flush()

	// Handle file reception
	fileReader := bufio.NewReader(fileShareStream)
	fileContents, err := io.ReadAll(fileReader)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	err = os.MkdirAll("received_files", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	err = os.WriteFile("received_file.png", fileContents, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("File received successfully")

}
