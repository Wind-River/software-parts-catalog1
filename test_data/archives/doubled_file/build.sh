#!/usr/bin/bash

_build () {
    cp src/date.txt src/date_copy.txt
    tar -cjvf doubled_file.tar.bz2 -C src date.txt date_copy.txt
}

_clean () {
    rm -f doubled_file.tar.bz2
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