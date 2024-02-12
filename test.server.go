package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	libp2p "github.com/libp2p/go-libp2p"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	// "github.com/libp2p/go-libp2p/p2p/security/noise"
)

var listenAddr = flag.String("listen", "/ip4/127.0.0.1/tcp/8080", "The address to listen on")

func main() {
	flag.Parse()

	node := initNode(*listenAddr)
	defer node.Close()

	peerInfo := getPeerInfo(node)
	combinedAddr, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		log.Fatalf("Error converting peer info to p2p addrs: %v", err)
	}

	setChatStreamHandler(node)
	setFileShareStreamHandler(node)

	log.Printf("Server is listening on %s", combinedAddr[0])

	waitForSignal(node)
}

// func initNode(listenAddr string) host.Host {
// 	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
// 	if err != nil {
// 		log.Fatalf("Error generating key pair: %v", err)
// 	}

// 	node, err := libp2p.New(libp2p.Identity(priv),
// 		libp2p.ListenAddrStrings(listenAddr),
// 	)
// 	if err != nil {
// 		log.Fatalf("Error creating libp2p node: %v", err)
// 	}
// 	return node
// }

func initNode(listenAddr string) host.Host {
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		log.Fatalf("Error generating key pair: %v", err)
	}

	// Create a Noise transport with the generated private key.
	// noiseTransport, err := noise.New(noise.ID, priv, nil)
	// if err != nil {
	// 	log.Fatalf("Error creating Noise transport: %v", err)
	// }

	node, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(listenAddr),
		// libp2p.Security(noise.ID, noiseTransport), // Add Noise transport
	)
	if err != nil {
		log.Fatalf("Error creating libp2p node: %v", err)
	}
	return node
}

func getPeerInfo(node host.Host) peer.AddrInfo {
	return peer.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
}

func setChatStreamHandler(node host.Host) {
	node.SetStreamHandler(protocol.ID("/chat/1.0.0"), func(s network.Stream) {
		fmt.Println("New stream opened")
		reader := bufio.NewReader(s)
		writer := bufio.NewWriter(s)

		for {
			// Read a message from the client
			message, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("Client closed the connection")
					break
				}
				fmt.Println("Error reading from client:", err)
				sendError(writer, "Error reading from client")
				break
			}
			fmt.Printf("Received message from client: %s\n", message)

			// Respond to the client
			_, err = writer.WriteString("Hello, client!\n")
			if err != nil {
				fmt.Println("Error writing to client:", err)
				sendError(writer, "Error writing to client")
				break
			}
			writer.Flush()
		}

		err := s.Close()
		if err != nil {
			fmt.Println("Error closing stream:", err)
		}
	})
}

const fileShareProtocolID = protocol.ID("/fileshare/1.0.0")

const chunkSize = 1024 // Size of each chunk in bytes

func setFileShareStreamHandler(node host.Host) {
	node.SetStreamHandler(fileShareProtocolID, func(s network.Stream) {
		fmt.Println("New file share stream opened")
		reader := bufio.NewReader(s)
		writer := bufio.NewWriter(s)

		// Read a message from the client
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed the connection")
				return
			}
			fmt.Println("Error reading from client:", err)
			sendError(writer, "Error reading from client")
			return
		}

		// Assume the message is the name of the file the client wants
		filename := strings.TrimSpace(message)

		// Open the file
		file, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error opening file:", err)
			sendError(writer, "Error opening file")
			return
		}
		defer file.Close()

		// Send the file contents to the client in chunks
		buf := make([]byte, chunkSize)
		for {
			n, err := file.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error reading file:", err)
				sendError(writer, "Error reading file")
				return
			}

			_, err = writer.Write(buf[:n])
			if err != nil {
				fmt.Println("Error sending file:", err)
				sendError(writer, "Error sending file")
				return
			}
			writer.Flush()
		}

		err = s.Close()
		if err != nil {
			fmt.Println("Error closing stream:", err)
		}
	})
}

func sendError(writer *bufio.Writer, errorMessage string) {
	_, err := writer.WriteString(errorMessage + "\n")
	if err != nil {
		fmt.Println("Error sending error message:", err)
	}
	writer.Flush()
}

func waitForSignal(node host.Host) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Println("Received signal, shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := node.Close(); err != nil {
		log.Fatalf("Error closing libp2p node: %v", err)
	}

	select {
	case <-ctx.Done():
		log.Println("Shutdown gracefully")
	case <-time.After(6 * time.Second):
		log.Println("Shutdown forcefully")
	}
}
