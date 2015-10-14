#include "proc_windows.h"

int add(int x, int y) {
	return x + y;
}

BOOL wait(DWORD* threadID) {
	DEBUG_EVENT debug_event = {0};
	for(;;) {
		if (!WaitForDebugEvent(&debug_event, INFINITE))
			return -1;
		switch(debug_event.dwDebugEventCode) {
			case CREATE_PROCESS_DEBUG_EVENT: 
				printf("Process created: hProcess=%d, hThread=%d, processID=%d, threadID=%d\n", 
					debug_event.u.CreateProcessInfo.hProcess, 
					debug_event.u.CreateProcessInfo.hThread, 
					debug_event.dwThreadId, 
					debug_event.dwProcessId);
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case LOAD_DLL_DEBUG_EVENT: 
				printf("Load DLL: %d\n", debug_event.u.LoadDll.lpBaseOfDll);
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case CREATE_THREAD_DEBUG_EVENT:
				printf("Create Thread: %d\n", debug_event.u.CreateThread.hThread);
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case UNLOAD_DLL_DEBUG_EVENT:
				printf("Unload DLL: %d\n", debug_event.u.UnloadDll.lpBaseOfDll);
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case EXCEPTION_DEBUG_EVENT:
				printf("Exception: %u, %d, #Params: %d, FirstChance: %d\n", 
					debug_event.u.Exception.ExceptionRecord.ExceptionCode,
					debug_event.u.Exception.ExceptionRecord.ExceptionFlags,
					debug_event.u.Exception.ExceptionRecord.NumberParameters,
					debug_event.u.Exception.dwFirstChance
				);
				for(int i = 0; i < debug_event.u.Exception.ExceptionRecord.NumberParameters; i++) {
					printf("\t %d\n", debug_event.u.Exception.ExceptionRecord.ExceptionInformation[i]);
				}
				break;
			case EXIT_THREAD_DEBUG_EVENT:
				printf("Exit thread: %d\n", debug_event.u.ExitThread.dwExitCode);
				ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
				continue;
			case EXIT_PROCESS_DEBUG_EVENT:
				printf("Exit process: %d\n", debug_event.u.ExitProcess.dwExitCode);
				*threadID = 0;
				return 0;
			default:
				printf("Unexpected.. %d\n", debug_event.dwDebugEventCode);
				return -1;	
		}
		printf("ThreadID: %d\n", debug_event.dwThreadId);
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
			printf("Process created: hProcess=%d, hThread=%d, processID=%d, threadID=%d\n", *hProcess, *hThread, processID, *threadID);
			
			DWORD dummyThreadID;
			ContinueDebugEvent(debug_event.dwProcessId, debug_event.dwThreadId, DBG_CONTINUE);
			return wait(&dummyThreadID);
		default:
			printf("Unexpected.. %d", debug_event.dwDebugEventCode);
			return -1;
	}
}