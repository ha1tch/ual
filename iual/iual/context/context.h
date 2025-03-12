#ifndef CONTEXT_H
#define CONTEXT_H

#include "stacks/stacks.h"
#include "spawn/spawn.h"

typedef struct SpawnNode {
    Spawn *sp;
    struct SpawnNode *next;
} SpawnNode;

typedef struct {
    IntStack *dstack;         // Default data stack
    IntStack *rstack;         // Return stack
    int globalMemory[1024];   // Global memory for store/load
    SpawnNode *spawnList;     // Linked list of spawns
} UalContext;

// Initialization: allocates and sets up a new UalContext.
UalContext *initContext();

// Cleanup: frees all resources in the context.
void freeContext(UalContext *ctx);

#endif
