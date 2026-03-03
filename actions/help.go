package actions

import (
	"fmt"
	"os"

	"bond/configuration"
)

func Help(commands []Command, configOptions []configuration.ConfigOption) {
	fmt.Printf("Usage: %s <command> [options]\n\n", os.Args[0])

	fmt.Printf("Commands:\n")
	for _, cmd := range commands {
		fmt.Printf("  %-8s %s\n", cmd.Name, cmd.Description)
	}
	fmt.Println()

	fmt.Printf("Options:\n")
	for _, option := range configOptions {
		fmt.Printf("  --%s", option.Name)

		if option.Type != "bool" {
			fmt.Printf(" %s", option.Type)
			fmt.Printf("\n      default: %s", option.Default)
		}

		fmt.Printf("\n      %s\n\n", option.Description)
	}
}
