#include "stacks.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Global memory for STORE/LOAD operations (for int stacks)
int globalMemory[1024] = {0};

// Define the global int stack list.
IntStackNode *globalIntStackList = NULL;

// Implementation of findIntStackPerspective().
IntStackPerspective *findIntStackPerspective(const char *name) {
    IntStackNode *cur = globalIntStackList;
    while (cur) {
        if (strcmp(cur->name, name) == 0)
            return cur->persp;
        cur = cur->next;
    }
    return NULL;
}
