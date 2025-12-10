#include "context.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Initialize a new context and its components.
UalContext *initContext() {
    UalContext *ctx = malloc(sizeof(UalContext));
    if (!ctx) {
        perror("malloc initContext");
        exit(EXIT_FAILURE);
    }
    // Initialize default stacks.
    ctx->dstack = newIntStack();
    ctx->rstack = newIntStack();
    // Initialize global memory to 0.
    memset(ctx->globalMemory, 0, sizeof(ctx->globalMemory));
    // Spawn list initially empty.
    ctx->spawnList = NULL;
    return ctx;
}

// Free all spawn nodes and join their threads if necessary.
static void freeSpawnList(SpawnNode *node) {
    while (node) {
        SpawnNode *next = node->next;
        // Optionally, join the spawn thread here if you want a graceful shutdown.
        // For example: pthread_join(node->sp->thread, NULL);
        // Free the spawn structure.
        free(node->sp->name);
        // (Assuming spawn->script is already freed during processing.)
        pthread_mutex_destroy(&node->sp->mtx);
        pthread_cond_destroy(&node->sp->cond);
        free(node->sp);
        free(node);
        node = next;
    }
}

// Cleanup context and free all resources.
void freeContext(UalContext *ctx) {
    if (ctx) {
        if (ctx->dstack) freeIntStack(ctx->dstack);
        if (ctx->rstack) freeIntStack(ctx->rstack);
        freeSpawnList(ctx->spawnList);
        free(ctx);
    }
}
