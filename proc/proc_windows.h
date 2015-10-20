#include <stdlib.h>
#include <stdio.h>
#include <sys/types.h>
#include <windows.h>
#include <tlhelp32.h>

BOOL wait(DWORD* threadID, DWORD* exitCode);
int waitForCreateProcessEvent(HANDLE* hProcess, HANDLE* hThread, int* threadID);
