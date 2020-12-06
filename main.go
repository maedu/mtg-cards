package main

import (
	"fmt"

	env "bitbucket.org/spinnerweb/accounting_common/env"
	cardApi "github.com/maedu/mtg-cards/card/api"
	"github.com/maedu/mtg-cards/draft/sealed"
	"github.com/maedu/mtg-cards/server"
)

func main() {
	server := server.Configure()
	cardApi.Setup(server)
	sealed.Setup(server)
	server.Run(fmt.Sprintf("0.0.0.0:%s", env.GetEnv("SERVER_PORT", "4004"))) // listen and serve on 0.0.0.0:8080
}
