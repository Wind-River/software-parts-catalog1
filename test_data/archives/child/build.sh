#!/usr/bin/bash

_build() {
    tar -cvjf src/bar.tar.bz2 -C src b.txt
    tar -cvjf child.tar.bz2 -C src a.txt bar.tar.bz2
}

_clean() {
    rm -f src/bar.tar.bz2
    rm -f child.tar.bz2
}

case $1 in
    "") # default to build
        _build
    ;;
    build)
        _build
    ;;
    clean)
        _clean
    ;;
    *)
        echo "Unexpected command '$1'"
        exit 1
esac