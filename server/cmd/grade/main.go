package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    Assignment string `help:"Path to assignment JSON files." required:"" type:"existingfile"`
    Submission string `help:"Path to submission directory." required:"" type:"existingdir"`
    OutPath string `help:"Option path to output a JSON grading result." type:"path"`
    User string `help:"User email for the submission." default:"testuser"`
}

func main() {
    kong.Parse(&args);
    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    assignment := model.MustLoadAssignmentConfig(args.Assignment);

    result, err := grader.GradeDefault(assignment, args.Submission, args.User);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to run grader.");
    }

    if (args.OutPath != "") {
        err = util.ToJSONFileIndent(result, args.OutPath, "", "    ");
        if (err != nil) {
            log.Fatal().Err(err).Str("outpath", args.OutPath).Msg("Failed to output JSON result.");
        }
    }

    fmt.Println(result.Report());
}
