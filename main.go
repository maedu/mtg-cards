package main

import (
	"fmt"

	env "bitbucket.org/spinnerweb/accounting_common/env"
	cardApi "bitbucket.org/spinnerweb/cards/card/api"
	game "bitbucket.org/spinnerweb/cards/draft/booster/game"
	"bitbucket.org/spinnerweb/cards/draft/sealed"
	"bitbucket.org/spinnerweb/cards/server"
)

func main() {
	server := server.Configure()
	cardApi.Setup(server)
	sealed.Setup(server)
	game.Setup(server)
	server.Run(fmt.Sprintf("0.0.0.0:%s", env.GetEnv("SERVER_PORT", "4004"))) // listen and serve on 0.0.0.0:8080
}
