package actions

import (
	"fmt"
	"os"

	"bond/configuration"
)

func Help(configOptions []configuration.ConfigOption) {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
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
