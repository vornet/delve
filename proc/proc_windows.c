#include "proc_windows.h"
 
int waitForCreateProcessEvent(HANDLE* hProcess, HANDLE* hThread, int* threadID) {
	DEBUG_EVENT debug_event = {0};
	if (!WaitForDebugEvent(&debug_event, INFINITE))
		return -1;
	switch(debug_event.dwDebugEventCode) {
		case CREATE_PROCESS_DEBUG_EVENT: 
			*hProcess = debug_event.u.CreateProcessInfo.hProcess; 
			*hThread = debug_event.u.CreateProcessInfo.hThread;
			*threadID = debug_event.dwThreadId;
			DWORD processID = debug_event.dwProcessId;
			
			return 0;
		default:
			return -1;
	}
}
