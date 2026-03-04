package main

import (
	"encoding/json"
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
	{Name: "serve", Description: "start the bond server (default)"},
	{Name: "sync", Description: "trigger a sync on the running bond instance"},
	{Name: "help", Description: "show this help"},
}

var configOptions = []configuration.ConfigOption{
	{
		Name:        "config",
		Default:     "~/.bond/bond.conf",
		Type:        "file-path",
		Description: "config file",
	},
	{
		Name:        "rest-listen-url",
		Default:     "0.0.0.0:80",
		Type:        "uint16",
		Description: "host:port for the public query REST API server to listen on",
	},
	{
		Name:        "control-listen-url",
		Default:     "127.0.0.1:8081",
		Type:        "string",
		Description: "host:port for the internal control REST API server to listen on",
	},
	{
		Name:        "ord-url",
		Default:     "http://localhost:4080",
		Type:        "url",
		Description: "host:port for the REST API on the ord server",
	},
	{
		Name:        "data-dir",
		Default:     "~/.bond/data",
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
	{
		Name:        "no-auto-index",
		Default:     "false",
		Type:        "bool",
		Description: "disable automatic indexing, only sync on demand via 'bond sync'",
	},
}

type Config struct {
	ConfigFile       string `config:"config"`
	RestListenUrl    string `config:"rest-listen-url"`
	ControlListenUrl string `config:"control-listen-url"`
	OrdUrl           string `config:"ord-url"`
	DataDir          string `config:"data-dir"`
	StartBlock       uint64 `config:"start-block"`
	NonInteractive   bool   `config:"non-interactive"`
	NoAutoIndex      bool   `config:"no-auto-index"`
}

func (c *Config) printConfig() {
	fmt.Printf("\nConfiguration:\n")

	val := reflect.ValueOf(c).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i).Name
		value := val.Field(i).Interface()
		fmt.Printf("  %-16s %v\n", field+":", value)
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
	queryServer := infrastructure.NewQueryHttpServer(nameResolver)
	controlServer := infrastructure.NewControlHttpServer(updater.SyncRequests())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go updater.Start(config.StartBlock)
	go queryServer.Start(config.RestListenUrl)
	go controlServer.Start(config.ControlListenUrl)

	<-sigChan
	fmt.Printf("\nShutting down gracefully...\n")

	if err := store.Close(); err != nil {
		fmt.Printf("Error closing database: %v\n", err)
	} else {
		fmt.Printf("Database closed successfully\n")
	}
}

func runSync() {
	config := configuration.ParseConfig[Config](configOptions)
	client := resty.New()
	client.SetTimeout(10 * time.Second)

	url := fmt.Sprintf("http://%s/sync", config.ControlListenUrl)

	resp, err := client.R().Post(url)
	if err != nil {
		fmt.Printf("Failed to call control API: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode() != 200 {
		fmt.Printf("Control API returned status %d: %s\n", resp.StatusCode(), resp.String())
		os.Exit(1)
	}

	var body struct {
		SyncedToBlock uint64 `json:"synced_to_block"`
	}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		fmt.Printf("Failed to parse control API response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Synced to block %d\n", body.SyncedToBlock)
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
	updater := actions.NewIndexUpdater(ord, btcNameStore, routingStore, blockStore, config.StartBlock, config.NonInteractive, config.NoAutoIndex)
	nameResolver := actions.NewNameResolver(btcNameStore, routingStore)
	return updater, nameResolver
}
