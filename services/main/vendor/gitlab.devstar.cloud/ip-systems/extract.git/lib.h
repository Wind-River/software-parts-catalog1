#include <stdbool.h>
#include "status.h"

void mustChdir(const char* path);
status extractOne(char* source, char* sourceName, char* dest, bool enclosed);