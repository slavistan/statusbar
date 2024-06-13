#!/usr/bin/env zsh

BIN_DIR_PATH=${ZSH_ARGZERO:A:h}/bin
BIN_NAME="todogrep"

_get_dlv_port() {
    if ! port=$(jq -r '.configurations[0].port' ${ZSH_ARGZERO:A:h}/.vscode/launch.json 2>/dev/null); then
        port=12345
    fi
    echo $port
}

_kill_dlv() {
    pkill -f "^dlv.*127\\.0\\.0\\.1:$(_get_dlv_port)"
}

test() {
    cd ${ZSH_ARGZERO:A:h}
    go test -v
    cd -
}

testd() {
    cd ${ZSH_ARGZERO:A:h}
    mkdir -p $BIN_DIR_PATH
    bin_path=$BIN_DIR_PATH/$BIN_NAME-debug.test
    go test -gcflags "-N -l" -c -o $bin_path

    # HACK: dlv ignores SIGINT from its executing shell
    #       but still can be killed via kill -SIGINT o_O
    trap '_kill_dlv' INT
    dlv exec --headless -l 127.0.0.1:$(_get_dlv_port) -- $bin_path &
    wait
    trap - INT

    cd -
}

run() {
    go run . $@
}

rund() {
    # Can be run from any directory.
    cd ${ZSH_ARGZERO:A:h}
    mkdir -p $BIN_DIR_PATH
    bin_path=$(realpath $BIN_DIR_PATH/$BIN_NAME-debug)
    go build -gcflags "-N -l" -o $bin_path
    cd -
    dlv exec --headless -l 127.0.0.1:$(_get_dlv_port) $bin_path -- $@
}

# Compiles project and tests/benchmarks. We provide a dedicated binary for the
# tests and benchmarks to be able to run them outside of the working tree.
#
# /path/to/todogrep.test -test.run=^$ -test.bench=BenchmarkName
#                        ^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^~~~ chooses a particular benchmark
#                        ^^^^^^^^^^^^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ disables tests
build() {
    mkdir -p $BIN_DIR_PATH
    cd ${ZSH_ARGZERO:A:h}
    go build -o $BIN_DIR_PATH/$BIN_NAME
    go test -c -o $BIN_DIR_PATH/$BIN_NAME.test
    cd -
}

clean() {
    rm -rf $BIN_DIR_PATH
}

if whence -w -- "$1" >/dev/null; then
    $@
else
    echo "Invalid command: $@" >&2
    exit 1
fi
