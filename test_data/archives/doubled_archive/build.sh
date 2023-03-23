#!/usr/bin/bash

_build () {
    tar -cvjf src/foo.tar.bz2 -C src b.txt
    tar -cvjf src/bar.tar.bz2 -C src b.txt
    tar -cvjf doubled_archive.tar.bz2 -C src a.txt foo.tar.bz2 bar.tar.bz2
}

_clean () {
    rm -f doubled_archive.tar.bz2 src/foo.tar.bz2 src/bar.tar.bz2
}

case $1 in
    '')
        _build # default to build if 0 arguments
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