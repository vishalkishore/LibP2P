package main

import (
	"bufio"
	"context"
	"fmt"
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

	info := getServerInfo("/ip4/127.0.0.1/tcp/8080/p2p/QmadYGN3eoVQD2A7U7pd1wT5St7pFcGcaFUjwZ7ojgqMoz")

	connectToServer(ctx, h, info)
	// if err != nil {
	// 	fmt.Println("Error connecting to server:", err)
	// 	return
	// }

	s := openStream(ctx, h, info, "/file-share")

	reader := bufio.NewReader(s)
	writer := bufio.NewWriter(s)

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
