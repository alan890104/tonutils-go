package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/alan890104/tonutils-go/address"
	"github.com/alan890104/tonutils-go/liteclient"
	"github.com/alan890104/tonutils-go/ton"
)

func main() {
	client := liteclient.NewConnectionPool()

	// connect to testnet lite server
	err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
	if err != nil {
		panic(err)
	}

	// initialize ton api lite connection wrapper
	api := ton.NewAPIClient(client)

	// if we want to route all requests to the same node, we can use it
	ctx := client.StickyContext(context.Background())

	b, err := api.LookupBlock(ctx, -1, -0x8000000000000000, 27531166)
	if err != nil {
		log.Fatalln("get block err:", err.Error())
		return
	}

	addr := address.MustParseAddr("Ef9VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVbxn")

	props, err := api.RunGetMethod(ctx, b, addr, "list_proposals")
	if err != nil {
		log.Fatalln("get proposals err:", err.Error())
		return
	}

	json.NewEncoder(os.Stdout).Encode(props.AsTuple())
}
