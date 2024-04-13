package main

import (
	"authentication-service/internal/app"
)

const configsDir = "configs"
const envDir = ".env"

func main() {
	app.Run(configsDir, envDir)
}
