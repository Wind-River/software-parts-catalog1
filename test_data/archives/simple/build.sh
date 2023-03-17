#!/usr/bin/bash

_build () {
    tar -cjvf simple.tar.bz2 -C src .
}

_clean () {
    rm -f simple.tar.bz2
}

case $1 in
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