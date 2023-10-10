package main

import (
    "fmt"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/report"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    CoursePath string `help:"Path to course JSON file." arg:"" type:"existingfile"`
    Email []string `help:"Email addresses to send the report to (as HTML)." short:"e"`
    HTML bool `help:"Output report as html." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Fetch the grades for a specific course from canvas."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    course := model.MustLoadCourseConfig(args.CoursePath);

    report, err := report.GetCourseScoringReport(course);
    if (err != nil) {
        log.Fatal().Err(err).Str("course", course.ID).Msg("Failed to get scoring report.");
    }

    if (args.HTML) {
        html, err := report.ToHTML();
        if (err != nil) {
            log.Fatal().Err(err).Str("course", course.ID).Msg("Failed to generate HTML scoring report.");
        }

        fmt.Println(html);
    } else {
        fmt.Println(util.MustToJSONIndent(report));
    }

    if (len(args.Email) > 0) {
        html, err := report.ToHTML();
        if (err != nil) {
            log.Fatal().Err(err).Str("course", course.ID).Msg("Failed to generate HTML scoring report.");
        }

        subject := fmt.Sprintf("Autograder Scoring Report for %s", course.DisplayName);

        err = email.Send(args.Email, subject, html, true);
        if (err != nil) {
            log.Fatal().Err(err).Str("course", course.ID).Msg("Failed to send scoring report email.");
        }
    }
}