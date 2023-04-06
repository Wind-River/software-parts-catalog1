#include "status.h"

//extract requires pwd to be the destination directory
//if filename is null, it is expected _extract will extract the archive without problem
//if filename is not null, it is expected that the archive is not a tar, so _decompress should be tried if _extract fails
status extract(char *filepath, char *filename);