package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
)

var args struct {
    config.ConfigArgs
}

func main() {
    kong.Parse(&args,
        kong.Description("Try and connect to the database specified via config and report the status."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    err = db.Open();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not connect to database.");
    }

    fmt.Println("Database connection successful.");
}
