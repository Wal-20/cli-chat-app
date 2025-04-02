package main

import (
    "github.com/Wal-20/cli-chat-app/internal/api"
    "github.com/Wal-20/cli-chat-app/internal/config"
    "log"
)

func main() {

    err := config.InitDB()
    if err != nil {
        log.Fatal("DB not initialized")
    }

    api.NewServer()

}

