#include <stdlib.h>
#include <unistd.h>
#include <stdbool.h>
#include <errno.h>
#include <sys/stat.h>
#include <string.h>
#include <linux/limits.h>

#include "lib.h"
#include "logger.h"
#include "extract.h"
#include "filename.h"
#include "sha1.h"

void mustChdir(const char* path) {
    int err = chdir(path);
    if (err != 0) {
        elog("[mustChdir(%s) %s]\n", path, strerror(errno));
    }
}

status extractOne(char* source, char* sourceName, char* dest, bool enclosed) {
    char* name = NULL;
    if(sourceName == NULL) {
        filename_ptr sourceFileName = parseFilename(source);
        if(sourceFileName->basename != NULL) {
            filename_ptr baseFileName = parseFilename(sourceFileName->basename);
            if(baseFileName->ext == NULL || strcmp(baseFileName->ext, ".tar") != 0) {
                vlog("get_filename(%s) -> %s\n", sourceFileName->basename, sourceFileName->name);
            }
            filename_free(baseFileName);
        }
        filename_free(sourceFileName);
    }

    if(dest != NULL) mustChdir(dest);
    if(enclosed) {
        char* hash = sha1(source);
        mkdir(hash, 0755);
        mustChdir(hash);
        free(hash);
    }

    status ret = NULL;
    if(name == NULL) {
        ret = extract(source, sourceName);
    } else {
        ret = extract(source, name);
        free(name);
    }

    ret->tag = malloc(PATH_MAX);
    ret->tag = getcwd(ret->tag, PATH_MAX);
    return ret;
}