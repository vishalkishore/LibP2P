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

const (
	ChatProtocolID      = "/chat/1.0.0"
	FileShareProtocolID = "/fileshare/1.0.0"
)

// createHost creates a new libp2p host.
func createHost() (host.Host, error) {
	return libp2p.New()
}

// getServerInfo creates a peer.AddrInfo from a multiaddress string.
func getServerInfo(maddrStr string) (*peer.AddrInfo, error) {
	maddr, err := ma.NewMultiaddr(maddrStr)
	if err != nil {
		return nil, err
	}

	return peer.AddrInfoFromP2pAddr(maddr)
}

// connectToServer connects the host to the server.
func connectToServer(ctx context.Context, h host.Host, info *peer.AddrInfo) error {
	return h.Connect(ctx, *info)
}

// openStream opens a new stream to the server with the given protocol ID.
func openStream(ctx context.Context, h host.Host, info *peer.AddrInfo, protocolID string) (network.Stream, error) {
	return h.NewStream(ctx, info.ID, protocol.ID(protocolID))
}

// readUserInput reads a line of user input from stdin.
func readUserInput(prompt string) (string, error) {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", scanner.Err()
	}
	return scanner.Text(), nil
}

// handleChatStream handles the chat stream.
func handleChatStream(ctx context.Context, h host.Host, info *peer.AddrInfo) error {
	chatStream, err := openStream(ctx, h, info, ChatProtocolID)
	if err != nil {
		return fmt.Errorf("error opening chat stream: %w", err)
	}

	reader := bufio.NewReader(chatStream)
	writer := bufio.NewWriter(chatStream)

	for {
		message, err := readUserInput("Enter a message to send to the server: ")
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		_, err = writer.WriteString(message + "\n")
		if err != nil {
			return fmt.Errorf("error writing to server: %w", err)
		}
		writer.Flush()

		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading from server: %w", err)
		}
		fmt.Printf("Received response from server: %s\n", response)

		confirmation, err := readUserInput("Do you want to continue (yes/no)? ")
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
		if confirmation != "yes" {
			break
		}
	}

	return nil
}

// handleFileShareStream handles the file share stream.
func handleFileShareStream(ctx context.Context, h host.Host, info *peer.AddrInfo) error {
	fileShareStream, err := openStream(ctx, h, info, FileShareProtocolID)
	if err != nil {
		return fmt.Errorf("error opening file share stream: %w", err)
	}

	filePath := "D:\\FTP\\repo\\LibP2P\\send.png\n"

	fileWriter := bufio.NewWriter(fileShareStream)
	_, err = fileWriter.WriteString(filePath + "\n")
	if err != nil {
		return fmt.Errorf("error sending file path: %w", err)
	}
	fileWriter.Flush()

	err = os.MkdirAll("received_files", 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	outFile, err := os.Create("received_files/received_file.png")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
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
			return fmt.Errorf("error reading file: %w", err)
		}

		_, err = outFile.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
	}

	fmt.Println("File received successfully")
	return nil
}

func main() {
	ctx := context.Background()

	h, err := createHost()
	if err != nil {
		fmt.Println("Error creating host:", err)
		return
	}
	defer h.Close()

	info, err := getServerInfo("/ip4/127.0.0.1/tcp/8080/p2p/QmcdzHmdS4RSnVWw5TxH6veT6YZNgM34bNfPTSARDbpBc1")

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
		err = handleChatStream(ctx, h, info)
	case 2:
		err = handleFileShareStream(ctx, h, info)
	default:
		fmt.Println("Invalid choice")
		return
	}

	if err != nil {
		fmt.Println("Error:", err)
	}
}
