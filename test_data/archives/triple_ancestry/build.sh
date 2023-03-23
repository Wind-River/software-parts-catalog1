#!/usr/bin/bash

_build() {
    tar -cvjf src/zap.tar.bz2 -C src c.txt
    tar -cvjf src/bar.tar.bz2 -C src b.txt zap.tar.bz2
    tar -cvjf triple_ancestry.tar.bz2 -C src a.txt bar.tar.bz2
}

_clean() {
    rm -f src/bar.tar.bz2
    rm -f src/bar.tar.bz2
    rm -f triple_ancestry.tar.bz2
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