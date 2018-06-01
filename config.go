package main

import "os"

func getConfig(name string) string {
    result := os.Getenv(name)

    if len(result) == 0 {
        panic(name)
    }

    return result
}
