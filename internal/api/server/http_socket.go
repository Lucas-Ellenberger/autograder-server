package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
)

var apiServer *http.Server

const API_SERVER_LOCK = "internal.api.server.API_SERVER_LOCK"

func runAPIServer(routes *[]core.Route) (err error) {
	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = errors.Join(err, fmt.Errorf("API server panicked: '%v'.", value))
	}()

	// Unlock API_SERVER_LOCK explicitly on each code path to ensure proper release regardless of the outcome.
	lockmanager.Lock(API_SERVER_LOCK)
	if apiServer != nil {
		lockmanager.Unlock(API_SERVER_LOCK)
		return fmt.Errorf("API server is already running.")
	}

	var port = config.WEB_PORT.Get()

	log.Info("API Server Started.", log.NewAttr("port", port))

	apiServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: core.GetRouteServer(routes),
	}

	lockmanager.Unlock(API_SERVER_LOCK)

	err = apiServer.ListenAndServe()
	if err == http.ErrServerClosed {
		// Set err to nil if the API server stopped due to a graceful shutdown.
		err = nil
	}

	if err != nil {
		log.Error("API server returned an error.", err)
	}

	log.Info("API Server Stopped.", log.NewAttr("port", port))

	return err
}

func stopAPIServer() {
	lockmanager.Lock(API_SERVER_LOCK)
	defer lockmanager.Unlock(API_SERVER_LOCK)

	if apiServer == nil {
		return
	}

	tempApiServer := apiServer
	apiServer = nil

	err := tempApiServer.Shutdown(context.Background())
	if err != nil {
		log.Error("Failed to stop the API server.", err)
	}
}
