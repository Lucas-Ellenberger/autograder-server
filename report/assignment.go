package report

import (
    "fmt"
    "time"

    "gonum.org/v1/gonum/stat"

    "github.com/edulinq/autograder/common"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/util"
)

const (
    OVERALL_NAME = "<Overall>"
)

type AssignmentScoringReport struct {
    AssignmentName string `json:"assignment-name"`
    NumberOfSubmissions int `json:"number-of-submissions"`
    LatestSubmission common.Timestamp `json:"latest-submission"`
    Questions []*ScoringReportQuestionStats `json:"questions"`
}

type ScoringReportQuestionStats struct {
    QuestionName string `json:"question-name"`

    Min float64 `json:"min"`
    Max float64 `json:"max"`
    Median float64 `json:"median"`
    Mean float64 `json:"mean"`
    StdDev float64 `json:"standard-deviation"`

    MinString string `json:"-"`
    MaxString string `json:"-"`
    MedianString string `json:"-"`
    MeanString string `json:"-"`
    StdDevString string `json:"-"`
}

type UserSubmissionStats struct {
    UserEmail string `json:"user-email"`
    NumberOfSubmissions int64 `json:"number-of-submissions"`
    FirstSubmission common.Timestamp `json:"first-submission"`
    LatestSubmission common.Timestamp `json:"latest-submission"`
}

type AssignmentSubmissionStats struct {
    AssignmentName string `json:"assignment-name"`
    Users []*UserSubmissionStats `json:"users"`
}

const DEFAULT_VALUE float64 = -1.0;

func GetAssignmentScoringReport(assignment *model.Assignment) (*AssignmentScoringReport, error) {
    questionNames, scores, lastSubmissionTime, err := fetchScores(assignment);
    if (err != nil) {
        return nil, err;
    }

    numSubmissions := 0;
    questions := make([]*ScoringReportQuestionStats, 0, len(questionNames));

    for _, questionName := range questionNames {
        min, max := util.MinMax(scores[questionName]);
        mean, stdDev := stat.MeanStdDev(scores[questionName], nil);
        median := util.Median(scores[questionName]);

        stats := &ScoringReportQuestionStats{
            QuestionName: questionName,
            Min: util.DefaultNaN(min, DEFAULT_VALUE),
            Max: util.DefaultNaN(max, DEFAULT_VALUE),
            Median: util.DefaultNaN(median, DEFAULT_VALUE),
            Mean: util.DefaultNaN(mean, DEFAULT_VALUE),
            StdDev: util.DefaultNaN(stdDev, DEFAULT_VALUE),

            MinString: fmt.Sprintf("%0.2f", min),
            MaxString: fmt.Sprintf("%0.2f", max),
            MedianString: fmt.Sprintf("%0.2f", median),
            MeanString: fmt.Sprintf("%0.2f", mean),
            StdDevString: fmt.Sprintf("%0.2f", stdDev),
        };

        questions = append(questions, stats);
        numSubmissions = len(scores[questionName]);
    }

    report := AssignmentScoringReport{
        AssignmentName: assignment.GetName(),
        NumberOfSubmissions: numSubmissions,
        LatestSubmission: common.TimestampFromTime(lastSubmissionTime),
        Questions: questions,
    };

    return &report, nil;
}

func fetchScores(assignment *model.Assignment) ([]string, map[string][]float64, time.Time, error) {
    results, err := db.GetRecentSubmissions(assignment, model.RoleStudent);
    if (err != nil) {
        return nil, nil, time.Time{}, fmt.Errorf("Failed to get recent submission results: '%w'.", err);
    }

    questionNames := make([]string, 0);
    scores := make(map[string][]float64);
    lastSubmissionTime := time.Time{};

    for _, result := range results {
        if (result == nil) {
            continue;
        }

        resultTime, err := result.GradingStartTime.Time();
        if (err != nil) {
            return nil, nil, time.Time{}, fmt.Errorf("Failed to get submission result time: '%w'.", err);
        }

        if (resultTime.After(lastSubmissionTime)) {
            lastSubmissionTime = resultTime;
        }

        if (len(questionNames) == 0) {
            for _, question := range result.Questions {
                questionNames = append(questionNames, question.Name);
                scores[question.Name] = make([]float64, 0);
            }

            questionNames = append(questionNames, OVERALL_NAME);
        }

        total := 0.0
        max_points := 0.0

        for _, question := range result.Questions {
            var score float64 = 0.0;
            if (!util.IsZero(question.MaxPoints)) {
                score = question.Score / question.MaxPoints;
            }

            scores[question.Name] = append(scores[question.Name], score);

            total += question.Score;
            max_points += question.MaxPoints;
        }

        total_score := 0.0;
        if (!util.IsZero(max_points)) {
            total_score = total / max_points;
        }


        scores[OVERALL_NAME] = append(scores[OVERALL_NAME], total_score);
    }

    return questionNames, scores, lastSubmissionTime, nil;
}

func GetScoringReportSubmissionStats(assignment *model.Assignment) (*AssignmentSubmissionStats, error) {
    submissionStats, err := fetchSubmissionStats(assignment);    
    if (err != nil) {
        return nil, err;
    }

    report := AssignmentSubmissionStats{
        AssignmentName: assignment.GetName(),
        Users: submissionStats,
    };

    return &report, nil;
}

func fetchSubmissionStats(assignment *model.Assignment) ([]*UserSubmissionStats, error) {
    // Get all users for this course.
    users, err := db.GetUsers(assignment.GetCourse());
    if (err != nil) {
        return nil, fmt.Errorf("Failed to get student emails for course: '%w'.", err);
    }

    submissionStats := make([]*UserSubmissionStats, 0);
    var numSubmissions int64;

    // Loop over every user in the course.
    for email, user := range users {
        // Skip users that are not students.
        if (user.Role != model.RoleStudent) {
            continue;
        }

        results, err := db.GetSubmissionAttempts(assignment, email);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to get recent submission results: '%w'.", err);
        }

        numSubmissions = 0;
        firstTime := time.Time{};
        lastTime := time.Time{};

        // Loop over every submission for the current user.
        for _, result := range results {
            if (result == nil) {
                continue;
            }

            resultTime, err := result.Info.GradingStartTime.Time();
            if (err != nil) {
                return nil, fmt.Errorf("Failed to get submission result time: '%w'.", err);
            }

            if (resultTime.Before(firstTime)) {
                firstTime = resultTime;
            }

            if (resultTime.After(lastTime)) {
                lastTime = resultTime;
            }

            numSubmissions += 1;
        }

        // Set the corresponding values for the current user.
        stats := UserSubmissionStats{
            UserEmail: email,
            NumberOfSubmissions: numSubmissions,
            FirstSubmission: common.TimestampFromTime(firstTime),
            LatestSubmission: common.TimestampFromTime(lastTime),
        };

        submissionStats = append(submissionStats, &stats);
    }

    return submissionStats, nil;
}
