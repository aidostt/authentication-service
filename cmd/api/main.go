package main

import (
	"authentication-service/internal/app/http"
)

const configsDir = "configs"
const envDir = ".env"

func main() {
	http.Run(configsDir, envDir)
}
