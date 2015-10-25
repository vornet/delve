#include "threads_windows.h"

typedef NTSTATUS (WINAPI *pNtQIT)(HANDLE, LONG, PVOID, ULONG, PULONG);

SIZE_T read_memory(HANDLE hProcess, void* addr, void *d, int len) {
	SIZE_T count;
	int ret = ReadProcessMemory(hProcess, addr, d, len, &count);
	if(ret == 0) {
		return -1;
	}
	return count;
}

SIZE_T write_memory(HANDLE hProcess, void* addr, void *d, int len) {
	SIZE_T count;
	int ret = WriteProcessMemory(hProcess, addr, d, len, &count);
	if(ret == 0) {
		return -1;
	}
	return count;
}

int thread_basic_information(HANDLE h, THREAD_BASIC_INFORMATION* addr) {
	pNtQIT NtQueryInformationThread = (pNtQIT)GetProcAddress(GetModuleHandle("ntdll.dll"), "NtQueryInformationThread");

    if(NtQueryInformationThread == NULL) 
        return 0;
		
	NTSTATUS status = NtQueryInformationThread(h,ThreadBasicInformation,addr,48, 0);
	return status;
}
