package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    libp2p "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/protocol"
)

func main() {
    node := initNode()
    defer node.Close()

    peerInfo := getPeerInfo(node)
    combinedAddr, err := peer.AddrInfoToP2pAddrs(&peerInfo)
    if err != nil {
        panic(err)
    }

    setStreamHandler(node)

    fmt.Println("Server is listening on", combinedAddr[0])

    waitForSignal()
}

func initNode() host.Host {
    priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
    if err != nil {
        log.Fatal("Error generating key pair: ", err)
    }

    node, err := libp2p.New(libp2p.Identity(priv),
        libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/8080"),
    )
    if err != nil {
        panic(err)
    }
    return node
}

func getPeerInfo(node host.Host) peer.AddrInfo {
    return peer.AddrInfo{
        ID:    node.ID(),
        Addrs: node.Addrs(),
    }
}

func setStreamHandler(node host.Host) {
    node.SetStreamHandler(protocol.ID("/file-share"), func(s network.Stream) {
        fmt.Println("New stream opened")

        _, err := s.Write([]byte("Hello, client!"))
        if err != nil {
            panic(err)
        }

        err = s.Close()
        if err != nil {
            panic(err)
        }
    })
}

func waitForSignal() {
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    <-ch

    fmt.Println("Received signal, shutting down...")
}