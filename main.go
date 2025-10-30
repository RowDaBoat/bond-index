package main

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"bond/actions"
	"bond/configuration"
	"bond/infrastructure"
	"bond/service"

	"github.com/go-resty/resty/v2"
)

var configOptions = []configuration.ConfigOption{
	{
		Name:        "config",
		Default:     "$HOME/.bond/bond.conf",
		Type:        "file-path",
		Description: "config file",
	},
	{
		Name:        "rest-listen-url",
		Default:     "0.0.0.0:80",
		Type:        "uint16",
		Description: "port for the REST API server to listen on",
	},
	{
		Name:        "ord-url",
		Default:     "http://localhost:4080",
		Type:        "url",
		Description: "host:port for the REST API on the ord server",
	},
	{
		Name:        "data-dir",
		Default:     "$HOME/.bond/data",
		Type:        "directory-path",
		Description: "directory for data storage",
	},
	{
		Name:        "start-ordinal",
		Default:     "18681",
		Type:        "uint64",
		Description: "ordinal id to start indexing from",
	},
	{
		Name:        "start-block",
		Default:     "775607",
		Type:        "uint64",
		Description: "block id to start indexing from",
	},
	{
		Name:        "help",
		Default:     "false",
		Type:        "bool",
		Description: "show this help",
	},
}

type Config struct {
	ConfigFile    string `config:"config"`
	RestListenUrl string `config:"rest-listen-url"`
	OrdUrl        string `config:"ord-url"`
	DataDir       string `config:"data-dir"`
	StartOrdinal  uint64 `config:"start-ordinal"`
	StartBlock    uint64 `config:"start-block"`
	Help          bool   `config:"help"`
}

func (c *Config) printConfig() {
	fmt.Printf("\nConfiguration:\n")

	val := reflect.ValueOf(c).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i).Name
		value := val.Field(i).Interface()
		fmt.Printf("  %-14s %v\n", field+":", value)
	}
	fmt.Println()
}

func main() {
	config := configuration.ParseConfig[Config](configOptions)
	config.printConfig()

	if config.Help {
		actions.Help(configOptions)
	} else {
		store, ord, btcNameStore, routingStore, blockStore := buildServices(config)
		updater, nameResolver := buildActions(config, ord, btcNameStore, routingStore, blockStore)
		server := infrastructure.NewHttpServer(nameResolver)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			if err := server.Start(config.RestListenUrl); err != nil {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}()

		go func() {
			updater.Update(config.StartBlock)
		}()

		<-sigChan
		fmt.Printf("\nShutting down gracefully...\n")
		if err := store.Close(); err != nil {
			fmt.Printf("Error closing database: %v\n", err)
		} else {
			fmt.Printf("Database closed successfully\n")
		}
	}
}

func buildServices(config Config) (*infrastructure.Store, service.Ord, service.BtcNameStore, service.RoutingStore, service.BlockStore) {
	var ord service.Ord
	var btcNameStore service.BtcNameStore
	var routingStore service.RoutingStore
	var blockStore service.BlockStore

	client := resty.New()
	client.SetTimeout(10 * time.Second)
	ord = infrastructure.NewOrd(client, config.OrdUrl)
	store := infrastructure.NewStore(config.DataDir)
	btcNameStore = infrastructure.NewBtcNameStore(store)
	routingStore = infrastructure.NewRoutingStore(store)
	blockStore = infrastructure.NewBlockStore(store)

	return store, ord, btcNameStore, routingStore, blockStore
}

func buildActions(
	config Config,
	ord service.Ord,
	btcNameStore service.BtcNameStore,
	routingStore service.RoutingStore,
	blockStore service.BlockStore,
) (
	*actions.IndexUpdater,
	*actions.NameResolver,
) {
	updater := actions.NewIndexUpdater(ord, btcNameStore, routingStore, blockStore, config.StartBlock)
	nameResolver := actions.NewNameResolver(btcNameStore, routingStore)
	return updater, nameResolver
}
