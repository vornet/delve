#include <stdlib.h>
#include <stdio.h>
#include <sys/types.h>
#include <windows.h>
#include <Winternl.h>

typedef struct THREAD_BASIC_INFORMATION
{
  NTSTATUS ExitStatus;
  PVOID TebBaseAddress;
  CLIENT_ID ClientId;
  ULONG_PTR AffinityMask;
  LONG Priority;
  LONG BasePriority;

} THREAD_BASIC_INFORMATION,*PTHREAD_BASIC_INFORMATION;

SIZE_T read_memory(HANDLE hProcess, void* addr, void *d, int len);
SIZE_T write_memory(HANDLE hProcess, void* addr, void *d, int len);
BOOL continue_debugger(DWORD processId, DWORD threadId);
int thread_basic_information(HANDLE h, PTHREAD_BASIC_INFORMATION addr);