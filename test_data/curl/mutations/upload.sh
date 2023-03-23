#!/usr/bin/bash
NAME=$1
FILEPATH=$2

OPERATIONS_1='{"query":"mutation upload($file: Upload!, $name: String) {\n  uploadArchive(file: $file, name: $name) {\n    extracted\n    archive {\n      sha256\n      name\n    }\n  }\n}","variables":{"name": "'
OPERATIONS_2='","file": null},"operationName":"upload"}'
OPERATIONS="${OPERATIONS_1}${NAME}${OPERATIONS_2}"

echo "OPERATIONS=${OPERATIONS}"

curl http://localhost/api/graphql \
    -F operations="${OPERATIONS}" \
    -F map='{ "0": ["variables.file"] }' \
    -F 0=@${FILEPATH}
