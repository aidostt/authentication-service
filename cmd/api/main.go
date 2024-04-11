package main

import (
	"authentication-service/internal/app/grpc"
)

const configsDir = "configs"
const envDir = ".env"

func main() {
	//http.Run(configsDir, envDir)
	grpc.Run(configsDir, envDir)
}
