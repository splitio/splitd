package main

import (
	"flag"
	"fmt"

	"github.com/splitio/splitd/splitio/conf"
	"gopkg.in/yaml.v3"
)

func main() {
	command := flag.String("command", "", "command to execute")
	flag.Parse()
	switch *command {
	case "gen-config-template":
		generateTemplateWithDefaults()
	default:
		fmt.Println("invalid command supplied")
	}
}

func generateTemplateWithDefaults() {
	var cfg conf.Config
	cfg.PopulateWithDefaults()

	raw, err := yaml.Marshal(cfg)
	mustNotFail(err)

	fmt.Println("# vi:ft=yaml") // vim type hint
	fmt.Println(string(raw))

}

func mustNotFail(err error) {
	if err != nil {
		panic(err.Error())
	}
}
