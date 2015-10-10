#include <stdlib.h>
#include <stdio.h>
#include <sys/types.h>
#include <windows.h>

SIZE_T read_memory(HANDLE hProcess, void* addr, void *d, int len);
