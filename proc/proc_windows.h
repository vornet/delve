#include <stdlib.h>
#include <stdio.h>
#include <sys/types.h>
#include <windows.h>
#include <tlhelp32.h>

int add(int x, int y);

BOOL wait(DWORD* threadID);

int waitForCreateProcessEvent(HANDLE* hProcess, HANDLE* hThread, int* threadID);
