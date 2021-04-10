package main

import (
	"fmt"

	env "bitbucket.org/spinnerweb/accounting_common/env"
	cardApi "github.com/maedu/mtg-cards/card/api"
	"github.com/maedu/mtg-cards/draft/sealed"
	edhrecApi "github.com/maedu/mtg-cards/edhrec/api"
	"github.com/maedu/mtg-cards/server"
	setApi "github.com/maedu/mtg-cards/set/api"
	userApi "github.com/maedu/mtg-cards/user/api"
)

func main() {
	server := server.Configure()
	cardApi.Setup(server)
	setApi.Setup(server)
	sealed.Setup(server)
	edhrecApi.Setup(server)
	userApi.Setup(server)
	server.Run(fmt.Sprintf("0.0.0.0:%s", env.GetEnv("SERVER_PORT", "4004"))) // listen and serve on 0.0.0.0:8080
}
