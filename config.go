package main

import (
    "log"
    "os"
)

func getConfig(name string) string {
    result := os.Getenv(name)

    if len(result) == 0 {
        log.Fatalf("Sorry, I cannot get the environment variable: %v", name)
    }

    return result
}
