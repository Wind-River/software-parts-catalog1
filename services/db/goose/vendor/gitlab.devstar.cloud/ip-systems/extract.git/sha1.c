#include <openssl/sha.h>
#include <stdio.h>
#include <stdlib.h>
#include "logger.h"
#include "hex.h"

char* sha1(const char* filepath) {
    vlog("sha1(%s)\n", filepath);
    FILE* file = fopen(filepath, "rb");

    SHA_CTX ctx;
    SHA1_Init(&ctx);

    unsigned char buf[1024];
    size_t read = 0;
    while((read = fread(buf, 1, 1024, file)) != 0) {
        SHA1_Update(&ctx, buf, read);
    }

    unsigned char* hash = malloc(SHA_DIGEST_LENGTH);
    SHA1_Final(hash, &ctx);
    fclose(file);

    char* ret = bytesToHex(hash, SHA_DIGEST_LENGTH);
    
    free(hash);
    return ret;
}