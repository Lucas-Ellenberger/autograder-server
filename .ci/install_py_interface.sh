#!/bin/bash

# Install a matching version of the Python interface.

readonly TEMP_DIR='/tmp/__autograder__/autograder-py'

function fetch_repo() {
    local branch=$1

    echo "Fetching Python interface repo."

    if [[ -d "${TEMP_DIR}" ]] ; then
        echo "Found existing repo '${TEMP_DIR}', skipping clone."
    else
        mkdir -p "$(dirname "${TEMP_DIR}")"
        git clone "${REPO_URL}" "${TEMP_DIR}"
        if [[ $? -ne 0 ]] ; then
            echo "ERROR: Failed to clone repo: '${REPO_URL}'."
            return 1
        fi
    fi

    cd "${TEMP_DIR}"

    git checkout "${branch}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to checkout target branch ('${branch}')."
        return 2
    fi

    git pull
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to pull."
        return 3
    fi

    return 0
}

function install_interface() {
    echo "Installing Python interface."

    cd "${TEMP_DIR}"

    ./install.sh
    return $?
}

function main() {
    if [[ $# -lt 1 || $# -gt 2 ]] ; then
        echo "USAGE: $0 repo_owner [branch]"
        exit 1
    fi

    local repo_owner=$1
    readonly REPO_URL="https://github.com/${repo_owner}/autograder-py.git"

    local branch=$(git branch --show-current)
    if [[ $# -eq 2 ]] ; then
        branch=$2
    fi

    trap exit SIGINT

    fetch_repo "${branch}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to fetch the Python interface repo."
        return 2
    fi

    install_interface
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to install Python interface."
        return 3
    fi

    echo "Sucessfully installed Python interface."

    return 0
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
