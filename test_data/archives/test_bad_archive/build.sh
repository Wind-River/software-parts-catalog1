#!/usr/bin/bash

_build() {
    tar -cvjf test_bad_archive.tar.bz2 -C src a.txt test/bad.tar.bz2
}

_clean() {
    rm -f test_bad_archive.tar.bz2
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
