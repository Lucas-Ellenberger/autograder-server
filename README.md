# Autograder Server

[![Main](https://github.com/edulinq/autograder-server/actions/workflows/main.yml/badge.svg)](https://github.com/edulinq/autograder-server/actions/workflows/main.yml)

A server for automatically (and in real-time) grading programming assignments.

The autograding effort is broken into three main parts:
 1. The autograding server that can accept student submissions, run the assignment graders, and provide feedback to students.
    This repository implements that server.
 2. The interfaces through which users can interact with the autograding server.
    Since the autograding server interacts via a REST API, the requirements for an interface are very low.
    Currently, the [autograder-py Python package](https://github.com/edulinq/autograder-py) is the only official interface package.
 3. Course and assignment configurations, including graders to score student submissions.
    These materials should generally be kept private (since they include grading information),
    but we have made a [sample course](https://github.com/edulinq/cse-cracks-course) available.

## Quick Links

 - [Autograder Server](https://github.com/edulinq/autograder-server)
 - [Autograder Python Interface](https://github.com/edulinq/autograder-py)
 - [Autograder Sample Course](https://github.com/edulinq/cse-cracks-course)

## Installation

This project uses Go 1.21.
Development and deployment of this project rely on POSIX systems (e.g., Linux, macOS, WSL).

All code that is not intended to be exported (used in packages outside of the autograder) is in the `internal` package/directory.
Since this is a server and not a library, that is the majority of the code.

By default, assignments are graded using Docker.
Therefore when grading functionality is used,
Docker should be installed on the machine and accessible to the current user without additional permissions.
Users without Docker can run the server without Docker (see below).

The project adheres to standard Go standards,
so the `go` tool can be used to build, test, manage, etc.
Additionally, the `scripts/build.sh` script is provided which will build all executables in this project into the `bin` directory.

```
./scripts/build.sh
```

All executable mains are kept in the `cmd` directory.
Each includes a usage and responds to the `--help` flag.

## Running Executables

Once built, all executables are available in the `bin` directory and can be run directly.
For development, these executables can also be run via `go run`, which will rebuild them if necessary before running:
```
go run cmd/version/main.go
```

## Configuration

This project uses configuration options to set the behavior of its executables.
All executables that use autograder resources use the same configuration infrastructure
and can therefore be configured the same way and with the same options.

To see all the available options,
either look in the [config/options.go](internal/config/options.go) file,
use the `cmd/list-options` executable.
```
./bin/list-options
```

Options can be set on the command-line using the `-c`/`--config` flag.
For example:
```
./bin/logs-example --config log.level=debug
```

Options can also be set using environmental variables by prefixing the option keys
with `AUTOGRADER__` and replacing any `.` with `__`.
For example option key `docker.disable` can be set by:
```
AUTOGRADER__DOCKER__DISABLE='true' ./scripts/run_tests.sh
```

### Directories

The primary directory the autograder will use for storing information is referred to as the "work directory",
and is set to `<base dir>/<instance name>`.
Most other paths are configured to be relative to the work directory.

The base directory is set through the `dirs.base` option,
and defaults to `$XDG_DATA_HOME`.

The instance name is a way to configuration a unique name for your autograder instance.
It can be set with the `instance.name` option,
and defaults to `autograder`.

### Loading Options

Configurations will be loaded in the following order (later options override earlier ones):
 0. The command-line options are checked for `BASE_DIR`.
 1. Load options from environmental variables.
 2. Options are loaded from `WORK_DIR/config` (config.json then secrets.json).
 3. Options are loaded from the current working directory (config.json then secrets.json).
 4. Options are loaded from any files specified with `--config-path` (ordered by appearance).
 5. Options are loaded from the command-line (`--config` / `-c`).

The base directory (`dirs.base`) can ONLY be set via the command-line or environmental variables.
This prevents cycles from the base directory changing and loading new options.

### Key Configuration Options

Here are several key configurations you should be aware of:

 - `instance.name` -- A name for this autograder instance.
 - `dirs.base` -- The "base" data directory for the autograder.
    Caches, databases, and other files will be stored here.
 - `server.backup.dir` -- The location that course backups will be saved to.
 - `log.level` -- The logging level. Should be one of ["trace", "debug", "info", "warn", "error", "fatal"].

## Preparing for Grading

Before the server is ready to grade student submissions,
you may have to take some steps depending on your server and assignments.

### Docker Grading

If using the standard Docker-based grading,
then you should take two steps before starting the server for grading.

First, ensure that any required Docker images are accessible for building.
The required images depend on what courses you are hosting,
but the default images live in the [autograder-docker](https://github.com/edulinq/autograder-docker) repository.

Second, you will want to pre-build your grader images.
This step is not required as the autograder will ensure that grader images are up-to-date before running a grader,
but this will ensure that the first student to submit is not stuck with a long wait.
Building images can be done using the `cmd/build-images` executable,
which will build images for all known assignments.

```
./bin/build-images
```

### Non-Docker Grading

When Docker is not available,
users can choose to run the server without Docker.
Non-Docker grading will only work when the default Python grader is used.
Also note that running the grader without Docker is a potential security risk
and should be avoided in production.

To disable docker, set the `docker.disable` config option to `true`.

The [Python autograder interface](https://github.com/edulinq/autograder-py) must be installed for the non-Docker grader to work:
```
pip install autograder-py
```

## Running the Server

The main server is available via the `cmd/server` executable.
```
./bin/server
```

The `web.port` config option can be used to set the port the server listens on:
```
./bin/server -c web.port=80
```

If you want to run on a privileged port as a non-root user,
we recommend using [setcap](https://man.archlinux.org/man/setcap.8).
The `scripts/setcap.sh` script will do this for you:
```
./scripts/setcap.sh
```

### Running the Server for Testing

You may want to run the server for testing/debugging,
e.g., if you are developing an interface to the server.
We recommend two additional changes to how you would normally run the server:
```
go run cmd/server/main.go --unit-testing
```

First, we ran the server using `go run`,
This will ensure that the server executable is up-to-date before running it.
Second we used the `--unit-testing` flag,
which will set some testing options, create a clean new database, and load the test courses (inside the `testdata directory).

Additionally, when running the server in `--unit-testing` mode,
most configs may get overwritten by the testing infrastructure but environmental variables will not get overwritten.
For more information about config options see the [Configuration section](#configuration) of this document.

## Running Tests

This repository comes with several types of tests.
All these tests are run in the CI,
and can also be run using the `./.ci/run_all.sh` script:
```
./.ci/run_all.sh
```

Users may also choose to run them individually.

### Base Tests

The base tests are created with Go's `testing` standard library package,
and can therefore be run using `go test`.
The `scripts/run_tests.sh` script runs `go test` for each package:
```
./scripts/run_tests.sh
```

### Running in a Docker Container

The autograder can also be run from a [Docker](https://www.docker.com/) container either
by building the image from source or pulling it from [Docker Hub](https://hub.docker.com/u/edulinq).

The autograder container requires two mounts.
The first is `/var/run/docker.sock`, the socket that the Docker daemon listens on.
The second is the autograder's temporary directory `/tmp/autograder-temp/`.
Both mounts rely on a POSIX system.

There are two versions of the autograder-server image.
The "slim" image is as small as possible and rebuilds artifacts from source each time.
The "prebuilt" image is larger and contains prebuilt binaries of each artifact (and thus runs faster).
To build the images, you can use the following commands from the repository's root directory:
```
docker build -f docker/slim/Dockerfile -t autograder-server-slim .
docker build -f docker/prebuilt/Dockerfile -t autograder-server-prebuilt .
```

Either image may be used for the following commands.
To use the existing images on Docker Hub,
you can use `edulinq/autograder-server-slim` or `edulinq/autograder-server-prebuilt`.

To run the container run:
```
docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -v /tmp/autograder-temp/:/tmp/autograder-temp autograder-server-prebuilt <command>
```

Where `<command>` can be any command form the `cmd` folder. For example it can be `version`:
```
docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -v /tmp/autograder-temp/:/tmp/autograder-temp autograder-server-prebuilt version
```

If you want to run the server, it could be useful to add the -p flag to the command, as shown below:
```
-p <host port>:<container port>
```

For example, you may use the following command (which uses the autograder's default port of 8080).:
```
docker run -it --rm -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock -v /tmp/autograder-temp/:/tmp/autograder-temp autograder-server-prebuilt server
```
