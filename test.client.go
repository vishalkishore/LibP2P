package main

import (
    "context"
    "fmt"
	"io/ioutil"

    libp2p "github.com/libp2p/go-libp2p"
    // "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/protocol"
    ma "github.com/multiformats/go-multiaddr"
)

func main() {
    ctx := context.Background()

    // Create a new libp2p Host that listens on a random TCP port
    h, err := libp2p.New()
    if err != nil {
        panic(err)
    }

    // Create a multiaddress for the server
    maddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/8080/p2p/QmQfnhjNdgF8AZrYf2f9354T4bnhXf1zUyjttVcxbUPC5X")
    if err != nil {
        panic(err)
    }

    // Parse the multiaddress to get the server's peer ID
    info, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        panic(err)
    }

    // Connect to the server
    err = h.Connect(ctx, *info)


if err != nil {
	panic(err)
}

// Open a new stream to the server with the "/my-protocol" protocol ID
s, err := h.NewStream(ctx, info.ID, protocol.ID("/my-protocol"))
if err != nil {
	panic(err)
}

// Read the message from the server
buf, err := ioutil.ReadAll(s)
if err != nil {
	panic(err)
}

    fmt.Printf("Received message from server: %s\n", string(buf))
}