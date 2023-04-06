#include <stdbool.h>

typedef struct filename_struct {
    char* name;
    char* basename;
    const char* ext;
    bool tar;
}* filename_ptr;

const char* getBasename( filename_ptr fp);
const char* getExtension( filename_ptr fp);
bool compressedBinary( filename_ptr fp );
filename_ptr parseFilename( char *str );
void filename_free( filename_ptr fp );
