#ifndef WR_STATUS
#define WR_STATUS

#include "warnings.h"

typedef struct exit_struct {
	int code;
	char* message;
	char* tag;
	warning_array* warnings;
}* status;

status report_status(int code, const char* message, const char* tag, warning_array* warnings);
status success(warning_array* wa);
void status_free(status stat);

#endif