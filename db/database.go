package db

import (
    "fmt"
    "sync"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db/pg"
    "github.com/eriq-augustine/autograder/db/sqlite"
    "github.com/eriq-augustine/autograder/usr"
)

var backend Backend;
var dbLock sync.Mutex;

type Backend interface {
    Close() error;
    EnsureTables() error;
    GetCourseUsers(courseID string) (map[string]*usr.User, error);
}

func Open() error {
    dbLock.Lock();
    defer dbLock.Unlock();

    if (backend != nil) {
        return nil;
    }

    var err error;
    if (config.DB_PG_URI.Get() == "") {
        backend, err = sqlite.Open();
    } else {
        backend, err = pg.Open();
    }

	if (err != nil) {
        return fmt.Errorf("Failed to open database: %w.", err);
	}

    return backend.EnsureTables();
}

func Close() error {
    dbLock.Lock();
    defer dbLock.Unlock();

    if (backend == nil) {
        return nil;
    }

    err := backend.Close();
    backend = nil;

    return err;
}

func GetCourseUsers(courseID string) (map[string]*usr.User, error) {
    return backend.GetCourseUsers(courseID);
}
