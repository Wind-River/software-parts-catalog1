#include <stdio.h>

int vlog(const char* format, ...);
int flog(FILE* stream, const char* format, ...);
int elog(const char* format, ...);
// int logger_verbose;
char* jsonEscape(const char* s);