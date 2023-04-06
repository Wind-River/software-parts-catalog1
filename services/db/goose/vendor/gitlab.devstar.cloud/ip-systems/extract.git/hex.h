#ifndef WR_HEX
#define WR_HEX

#include <unistd.h>

//expects unsigned char (0~255)
char* bytesToHex(const unsigned char* bytes, size_t len);

//expects signed char (-128~127)
char* charToHex(const char* s, size_t len);

#endif