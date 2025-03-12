#include "context.h"
#include "stacks/stacks.h"
#include "spawn/spawn.h"
#include "interpreter/interpreter.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Global unified context pointer.
UalContext *ualCtx = NULL;

int main(void) {
    printf("iual v0.0.1\n");
    printf("iual is an exceedingly trivial interactive ual 0.0.1 interpreter\n");

    // Initialize the unified context.
    ualCtx = initContext();
    
    // Create default perspectives for the two int stacks.
    // "dstack" and "rstack" are the default names.
    IntStackPerspective *dsp = newIntStackPerspective(ualCtx->dstack, "dstack");
    IntStackPerspective *rsp = newIntStackPerspective(ualCtx->rstack, "rstack");
    
    // Create default spawn "spawn" and add it to the spawn list.
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
        line[strcspn(line, "\n")] = '\0';
        if (strlen(line) == 0)
            continue;

        // If the line begins with '@' and contains a colon, treat it as a compound command.
        if (line[0] == '@' && strchr(line, ':')) {
            processCompoundCommand(line);
            continue;
        }

        // Global command processing (very simplified demo).
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
                // For demonstration, we recreate the default data stack.
                if (ualCtx->dstack) {
                    freeIntStack(ualCtx->dstack);
                }
                ualCtx->dstack = newIntStack();
                dsp = newIntStackPerspective(ualCtx->dstack, name);
                printf("Created new int stack '%s'\n", name);
            } else if (strcmp(type, "str") == 0) {
                // String stacks not integrated in context yet.
                printf("String stacks not integrated in context yet.\n");
            } else {
                printf("Only int and str stacks are supported in this demo.\n");
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
        } else if (strcmp(cmd, "quit") == 0) {
            break;
        } else {
            printf("Unknown global command: %s\n", cmd);
        }
    }
    
    freeContext(ualCtx);
    return 0;
}
