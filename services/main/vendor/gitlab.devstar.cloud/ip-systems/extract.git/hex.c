#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <string.h>

char* bytesToHex(const unsigned char* bytes, size_t len) {
    char* ret = malloc(len*2+1);
    for(int i = 0; i < len; i++) {
        int j = i*2;
        sprintf(ret+j, "%02x", bytes[i]);
    }
    ret[len*2] = '\0';

    return ret;
}

char* charToHex(const char* s, size_t len) {
    if(len == 0) {
        len = strlen(s);
    }

    char* ret = malloc(len*2+1);
    for(int i = 0; i < len; i++) {
        int j = i*2;
        sprintf(ret+j, "%02x", s[i]);
        printf("%c -> %02x\n", s[i], s[i]);
    }
    ret[len*2] = '\0';

    return ret;
}