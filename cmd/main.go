package main

import cli "github.com/anzboi/protoc-gen-cli/test/proto"

func main() {
	if err := cli.AccountsListAccountCommand().Execute(); err != nil {
		panic(err)
	}
}
