#include "threads_windows.h"

SIZE_T read_memory(HANDLE hProcess, void* addr, void *d, int len) {
	SIZE_T count;
	int ret = ReadProcessMemory(hProcess, addr, d, len, &count);
	if(ret == 0) {
		return -1;
	}
	return count;
}
