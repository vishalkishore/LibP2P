package main

import (
    "context"
    "fmt"
    "io/ioutil"

    libp2p "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/protocol"
    ma "github.com/multiformats/go-multiaddr"
)

func main() {
    ctx := context.Background()

    h := createHost()
    defer h.Close()

    info := getServerInfo("/ip4/127.0.0.1/tcp/8080/p2p/QmeBc2W7QSAoFs4gPGVG3hy1gL82ixg1qHBxTYj6Z39RT1")

    connectToServer(ctx, h, info)

    s := openStream(ctx, h, info, "/file-share")

    message := readMessage(s)

    fmt.Printf("Received message from server: %s\n", string(message))
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

func readMessage(s network.Stream) []byte {
    buf, err := ioutil.ReadAll(s)
    if err != nil {
        panic(err)
    }
    return buf
}