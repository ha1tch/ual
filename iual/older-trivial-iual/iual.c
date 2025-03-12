#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "context.h"               // Defines UalContext, initContext, freeContext
#include "stacks/stacks.h"         // Defines IntStack, newIntStackPerspective, etc.
#include "spawn/spawn.h"           // For spawn operations
#include "interpreter/interpreter.h" // For processCompoundCommand()

// Global pointer for the unified context.
UalContext *ualCtx = NULL;

// Global linked list node for int stack perspectives.
typedef struct IntStackNode {
    char *name;
    IntStackPerspective *persp;
    struct IntStackNode *next;
} IntStackNode;

IntStackNode *globalIntStackList = NULL;

// Helper to add a new int stack perspective to the global list.
void addIntStackPerspective(const char *name, IntStackPerspective *persp) {
    IntStackNode *node = malloc(sizeof(IntStackNode));
    if (!node) {
        perror("malloc addIntStackPerspective");
        exit(EXIT_FAILURE);
    }
    node->name = strdup(name);
    node->persp = persp;
    node->next = globalIntStackList;
    globalIntStackList = node;
}

int main(void) {
    // Print header information.
    printf("iual v0.0.1\n");
    printf("iual is an exceedingly trivial interactive ual 0.0.1 interpreter\n");

    // Initialize the unified context.
    // This creates default stacks: dstack and rstack.
    ualCtx = initContext();
    
    // Wrap default stacks in perspective objects.
    IntStackPerspective *dsp = newIntStackPerspective(ualCtx->dstack);
    IntStackPerspective *rsp = newIntStackPerspective(ualCtx->rstack);
    
    // Add default stacks to the global list.
    addIntStackPerspective("dstack", dsp);
    addIntStackPerspective("rstack", rsp);

    // Create default spawn "spawn" and add it to the context's spawn list.
    Spawn *defaultSpawn = newSpawn("spawn");
    SpawnNode *node = malloc(sizeof(SpawnNode));
    if (!node) {
        perror("malloc spawn node");
        exit(EXIT_FAILURE);
    }
    node->sp = defaultSpawn;
    node->next = ualCtx->spawnList;
    ualCtx->spawnList = node;
    printf("Added spawn '%s'\n", defaultSpawn->name);

    // Main interactive loop.
    char line[256];
    while (1) {
        printf("> ");
        if (!fgets(line, sizeof(line), stdin))
            break;
        // Remove trailing newline.
        line[strcspn(line, "\n")] = '\0';
        if (strlen(line) == 0)
            continue;

        // Process compound commands: lines that begin with '@' and contain a colon.
        if (line[0] == '@' && strchr(line, ':')) {
            processCompoundCommand(line);
            continue;
        }

        // Global command processing:
        char *cmd = strtok(line, " ");
        if (!cmd)
            continue;

        if (strcmp(cmd, "new") == 0) {
            // Command: new <stack name> <int|str|float>
            char *name = strtok(NULL, " ");
            char *type = strtok(NULL, " ");
            if (!name || !type) {
                printf("Usage: new <stack name> <int|str|float>\n");
                continue;
            }
            if (strcmp(type, "int") == 0) {
                // Create a new int stack and its perspective.
                IntStack *newStack = newIntStack();
                IntStackPerspective *newPersp = newIntStackPerspective(newStack);
                addIntStackPerspective(name, newPersp);
                printf("Created new int stack '%s'\n", name);
            } else if (strcmp(type, "str") == 0) {
                printf("String stacks not integrated in context yet.\n");
            } else {
                printf("Unknown stack type: %s\n", type);
            }
        } else if (strcmp(cmd, "spawn") == 0) {
            // Command: spawn <goroutine name>
            char *name = strtok(NULL, " ");
            if (!name) {
                printf("Usage: spawn <goroutine name>\n");
                continue;
            }
            Spawn *sp = newSpawn(name);
            SpawnNode *newNode = malloc(sizeof(SpawnNode));
            if (!newNode) {
                perror("malloc spawn node");
                exit(EXIT_FAILURE);
            }
            newNode->sp = sp;
            newNode->next = ualCtx->spawnList;
            ualCtx->spawnList = newNode;
            printf("Added spawn '%s'\n", sp->name);
        } else if (strcmp(cmd, "list") == 0) {
            // Command: list (lists spawns)
            printf("Spawns:\n");
            SpawnNode *cur = ualCtx->spawnList;
            while (cur) {
                printf("  %s\n", cur->sp->name);
                cur = cur->next;
            }
            // Also list available int stacks.
            printf("Int Stacks:\n");
            IntStackNode *curStack = globalIntStackList;
            while (curStack) {
                printf("  %s\n", curStack->name);
                curStack = curStack->next;
            }
        } else if (strcmp(cmd, "quit") == 0) {
            break;
        } else {
            printf("Unknown global command: %s\n", cmd);
        }
    }
    
    // Cleanup: free context (which frees default stacks, etc.)
    freeContext(ualCtx);
    // (For brevity, not freeing the globalIntStackList in this demo.)
    return 0;
}
