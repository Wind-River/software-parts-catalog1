#ifndef WR_WARN
#define WR_WARN

#include <unistd.h>
#include <stdio.h>

typedef struct warn_struct {
    int line;
    char* file;
    char* message;
} warning;

typedef struct warn_array {
    size_t nwarn;
    warning** warnings;
} warning_array;

warning_array* warning_array_init();
void warn(warning_array* wa, int line, const char* file, const char* message);
void warning_array_free(warning_array* wa);
void warning_fprint(FILE* f, warning* w);

#endif