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

var commands = []actions.Command{
	{Name: "server", Description: "start the bond server (default)"},
	{Name: "sync", Description: "trigger a sync on the running bond instance"},
	{Name: "help", Description: "show this help"},
}

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
		Name:        "non-interactive",
		Default:     "false",
		Type:        "bool",
		Description: "disable interactive output",
	},
}

type Config struct {
	ConfigFile     string `config:"config"`
	RestListenUrl  string `config:"rest-listen-url"`
	OrdUrl         string `config:"ord-url"`
	DataDir        string `config:"data-dir"`
	StartBlock     uint64 `config:"start-block"`
	NonInteractive bool   `config:"non-interactive"`
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
	cmd := "serve"
	if len(os.Args) > 1 && !isFlag(os.Args[1]) {
		cmd = os.Args[1]
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	switch cmd {
	case "serve":
		runServer()
	case "sync":
		runSync()
	case "help":
		runHelp()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		runHelp()
	}
}

func isFlag(arg string) bool {
	return len(arg) > 1 && arg[0] == '-'
}

func runServer() {
	config := configuration.ParseConfig[Config](configOptions)
	config.printConfig()

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

func runSync() {
	fmt.Println("sync: not yet implemented")
}

func runHelp() {
	actions.Help(commands, configOptions)
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
	updater := actions.NewIndexUpdater(ord, btcNameStore, routingStore, blockStore, config.StartBlock, config.NonInteractive)
	nameResolver := actions.NewNameResolver(btcNameStore, routingStore)
	return updater, nameResolver
}
