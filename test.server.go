
package main

import (
	"fmt"
	"log"
	"os"
	// "context"
	"os/signal"
	"syscall"
	libp2p "github.com/libp2p/go-libp2p"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/network"

	// "github.com/multiformats/go-multiaddr"
	// "github.com/vishalkishore/p2pFS"
	"github.com/libp2p/go-libp2p-core/crypto"
)

func main() {
	node := initNode()
	
	peerInfo := GetPeerInfo(node)

	combinedAddr, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}

	fmt.Println("libp2p node address:", combinedAddr[0])

	// Set a stream handler on "/my-protocol"
    node.SetStreamHandler(protocol.ID("/my-protocol"), func(s network.Stream) {
        fmt.Println("New stream opened")

        // Send a message to the client
        _, err := s.Write([]byte("Hello, client!"))
        if err != nil {
            panic(err)
        }

        // Close the stream
        err = s.Close()
        if err != nil {
            panic(err)
        }
    })

    fmt.Println("Server is listening on", node.Addrs())

	WaitForSignal(node)
}

func initNode() (node host.Host) {
	// start a libp2p node with default settings
	// ctx := context.Background()

	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		log.Fatal("Error generating key pair: ", err)
	}

	node, err = libp2p.New(libp2p.Identity(priv),
			libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/8080"),
		)
	if err != nil {
		panic(err)
	}
	return node
}

func GetPeerInfo(node host.Host) (peerInfo peerstore.AddrInfo) {
	peerInfo = peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	return peerInfo
}

func WaitForSignal(node host.Host) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	fmt.Println("Received signal, shutting down...")

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}
