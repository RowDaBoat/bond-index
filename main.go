package main

import (
	"fmt"
	"reflect"
	"time"

	"yarr/actions"
	"yarr/configuration"
	"yarr/infrastructure"
	"yarr/infrastructure/memory"
	"yarr/service"

	"github.com/go-resty/resty/v2"
)

var configOptions = []configuration.ConfigOption{
	{
		Name:        "config",
		Default:     "$HOME/.yarr/yarr.conf",
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
		Default:     "localhost:4080",
		Type:        "url",
		Description: "host:port for the REST API on the ord server",
	},
	{
		Name:        "data-dir",
		Default:     "$HOME/.yarr/data",
		Type:        "directory-path",
		Description: "directory for data storage",
	},
	{
		Name:        "test",
		Default:     "false",
		Type:        "bool",
		Description: "run in test mode",
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
	TestMode      bool   `config:"test"`
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
		fmt.Printf("  %-12s %v\n", field+":", value)
	}
	fmt.Println()
}

func main() {
	config := configuration.ParseConfig[Config](configOptions)
	config.printConfig()

	if config.Help {
		actions.Help(configOptions)
	} else {
		ord, btcNameStore, routingStore, blockStore := buildServices(config)
		updater, nameResolver := buildActions(config, ord, btcNameStore, routingStore, blockStore)
		server := infrastructure.NewHttpServer(nameResolver)

		go func() {
			if err := server.Start(config.RestListenUrl); err != nil {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}()

		updater.Update(config.StartBlock)
	}
}

func buildServices(config Config) (service.Ord, service.BtcNameStore, service.RoutingStore, service.BlockStore) {
	var ord service.Ord
	var btcNameStore service.BtcNameStore
	var routingStore service.RoutingStore
	var blockStore service.BlockStore

	if config.TestMode {
		ord = memory.NewOrd(config.DataDir)
		btcNameStore = memory.NewBtcNameStore()
		routingStore = memory.NewRoutingStore()
		blockStore = memory.NewBlockStore()
	} else {
		client := resty.New()
		client.SetTimeout(10 * time.Second)
		ord = infrastructure.NewOrd(client, config.OrdUrl)
		store := infrastructure.NewStore(config.DataDir)
		btcNameStore = infrastructure.NewBtcNameStore(store)
		routingStore = infrastructure.NewRoutingStore(store)
		blockStore = infrastructure.NewBlockStore(store)
	}
	return ord, btcNameStore, routingStore, blockStore
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
