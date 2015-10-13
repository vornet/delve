#include "proc_windows.h"

int add(int x, int y) {
	return x + y;
}

void wait() {
	DEBUG_EVENT debug_event = {0};
	if (!WaitForDebugEvent(&debug_event, INFINITE))
		return;
}

int waitForCreateProcessEvent(HANDLE* hProcess, HANDLE* hThread, int* threadID) {
	DEBUG_EVENT debug_event = {0};
	if (!WaitForDebugEvent(&debug_event, INFINITE))
		return -1;
	switch(debug_event.dwDebugEventCode) {
		case CREATE_PROCESS_DEBUG_EVENT: 
			*hProcess = debug_event.u.CreateProcessInfo.hProcess; 
			*hThread = debug_event.u.CreateProcessInfo.hThread;
			*threadID = GetThreadId(*hThread);
			printf("Process created: hProcess=%d, hThread=%d\n, threadID=%d\n", *hProcess, *hThread, *threadID);
			return 0;
		default:
			printf("Unexpected.. %d", debug_event.dwDebugEventCode);
			return -1;
	}
}