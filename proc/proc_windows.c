#include "proc_windows.h"
 
BOOL wait(DWORD* threadID, DWORD* exitCode) {
	DEBUG_EVENT debug_event = {0};
	for(;;) {
		if (!WaitForDebugEvent(&debug_event, INFINITE))
			return -1;
		// TODO: This switch should be handled in GO so that events can
		// easily be processed (addThread, handle exit gracefully, etc.)
		switch(debug_event.dwDebugEventCode) {
			case CREATE_PROCESS_DEBUG_EVENT:
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case LOAD_DLL_DEBUG_EVENT: 
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case CREATE_THREAD_DEBUG_EVENT:
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case UNLOAD_DLL_DEBUG_EVENT:
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case EXCEPTION_DEBUG_EVENT:
				break;
			case EXIT_THREAD_DEBUG_EVENT:
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case EXIT_PROCESS_DEBUG_EVENT:
				*threadID = 0;
				*exitCode = debug_event.u.ExitProcess.dwExitCode;
				return 0;
			default:
				return -1;	
		}
		*threadID = debug_event.dwThreadId;
		return 0;
	}
}

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
			
			DWORD dummyThreadID;
			DWORD dummyExitCode;
			ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
			return wait(&dummyThreadID, &dummyExitCode);
		default:
			return -1;
	}
}
