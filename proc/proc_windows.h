#include <stdlib.h>
#include <stdio.h>
#include <sys/types.h>
#include <windows.h>
#include <tlhelp32.h>

int add(int x, int y);

void wait();

int waitForCreateProcessEvent(HANDLE* hProcess, int* hThread);
