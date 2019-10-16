package main

import (
    "os"
    "regexp"
)

func getConfig(name string) string {
    result := os.Getenv(name)

    if len(result) == 0 {
        panic(name)
    }

    return result
}

func getConfigExpression(name string) *regexp.Regexp {
    result := os.Getenv(name)

    if len(result) == 0 {
        panic(name)
    }

    return regexp.MustCompile(result)
}
