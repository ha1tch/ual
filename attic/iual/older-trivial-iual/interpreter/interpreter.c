#include "interpreter/interpreter.h"
#include "stacks/stacks.h"
#include "spawn/spawn.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

// Global linked list node for int stack perspectives is defined in iual.c.
typedef struct IntStackNode {
    char *name;
    IntStackPerspective *persp;
    struct IntStackNode *next;
} IntStackNode;
extern IntStackNode *globalIntStackList;

// Helper: find the perspective by stack name.
IntStackPerspective *findIntStackPerspective(const char *name) {
    IntStackNode *cur = globalIntStackList;
    while (cur) {
        if (strcmp(cur->name, name) == 0)
            return cur->persp;
        cur = cur->next;
    }
    return NULL;
}

// Helper: trim leading and trailing whitespace.
static char *trimWhitespace(char *str) {
    // Trim leading whitespace.
    while (isspace(*str)) str++;
    // Trim trailing whitespace.
    char *end = str + strlen(str) - 1;
    while (end > str && isspace(*end)) { *end = '\0'; end--; }
    return str;
}

// Helper: convert string to uppercase in a temporary buffer.
static void toUppercase(const char *src, char *dst, size_t dstSize) {
    size_t i;
    for (i = 0; i < dstSize - 1 && src[i]; i++) {
        dst[i] = toupper(src[i]);
    }
    dst[i] = '\0';
}

void processCompoundCommand(char *input) {
    // Compound command format: "@<selector>: <command tokens>"
    char *colon = strchr(input, ':');
    if (!colon) {
        printf("Invalid compound command format.\n");
        return;
    }
    *colon = '\0';
    char *selector = input + 1;  // Skip '@'
    char *commands = colon + 1;
    commands = trimWhitespace(commands);

    // Look up the int stack perspective by name.
    IntStackPerspective *stackPerspective = findIntStackPerspective(selector);
    if (!stackPerspective) {
        printf("No int stack named '%s' found.\n", selector);
        return;
    }
    
    char *token = strtok(commands, " ");
    while (token != NULL) {
        // Handle the bring operation specially.
        if (strncmp(token, "bring", 5) == 0) {
            char srcType[32], srcStackName[32];
            // Check for parentheses form.
            if (token[5] == '(' && token[strlen(token)-1] == ')') {
                // Extract substring inside the parentheses.
                char argBuffer[64];
                int argLen = strlen(token) - 6; // Exclude "bring(" and ")"
                if (argLen >= sizeof(argBuffer))
                    argLen = sizeof(argBuffer) - 1;
                strncpy(argBuffer, token + 6, argLen);
                argBuffer[argLen] = '\0';
                // Trim whitespace.
                char *args = trimWhitespace(argBuffer);
                // Now split by comma.
                char *param = strtok(args, ",");
                if (!param) {
                    printf("bring requires two parameters: srcType and srcStack\n");
                    token = strtok(NULL, " ");
                    continue;
                }
                strncpy(srcType, trimWhitespace(param), 31);
                srcType[31] = '\0';
                param = strtok(NULL, ",");
                if (!param) {
                    printf("bring requires two parameters: srcType and srcStack\n");
                    token = strtok(NULL, " ");
                    continue;
                }
                strncpy(srcStackName, trimWhitespace(param), 31);
                srcStackName[31] = '\0';
            } else {
                // Else, assume colon form: "bring:<srcType>,<srcStack>"
                char *colonPos = strchr(token, ':');
                if (!colonPos) {
                    printf("bring requires argument in form <srcType>,<srcStack>\n");
                    token = strtok(NULL, " ");
                    continue;
                }
                char argStr[64];
                strcpy(argStr, colonPos + 1);
                char *param = strtok(argStr, ",");
                if (!param) {
                    printf("bring requires two parameters: srcType and srcStack\n");
                    token = strtok(NULL, " ");
                    continue;
                }
                strncpy(srcType, trimWhitespace(param), 31);
                srcType[31] = '\0';
                param = strtok(NULL, ",");
                if (!param) {
                    printf("bring requires two parameters: srcType and srcStack\n");
                    token = strtok(NULL, " ");
                    continue;
                }
                strncpy(srcStackName, trimWhitespace(param), 31);
                srcStackName[31] = '\0';
            }
            // Remove any leading '@' from the source stack name.
            if (srcStackName[0] == '@') {
                memmove(srcStackName, srcStackName + 1, strlen(srcStackName));
            }
            IntStackPerspective *srcPersp = findIntStackPerspective(srcStackName);
            if (!srcPersp) {
                printf("No int stack named '%s' found.\n", srcStackName);
                token = strtok(NULL, " ");
                continue;
            }
            int ok;
            int value = intStackPerspectivePop(srcPersp, &ok);
            if (!ok) {
                printf("Source stack '%s' is empty.\n", srcStackName);
                token = strtok(NULL, " ");
                continue;
            }
            intStackPerspectivePush(stackPerspective, value);
            printf("Brought value from int stack '%s' to selected stack '%s'\n", srcStackName, selector);
            token = strtok(NULL, " ");
            continue;
        }
        
        // Handle function-like syntax for non-bring commands.
        if (strchr(token, '(') && token[strlen(token) - 1] == ')') {
            char *paren = strchr(token, '(');
            *paren = '\0';
            char *op = token;
            char *argList = paren + 1;
            argList[strlen(argList) - 1] = '\0';
            char *arg = strtok(argList, ",");
            while (arg != NULL) {
                arg = trimWhitespace(arg);
                intStackPerspectivePush(stackPerspective, atoi(arg));
                arg = strtok(NULL, ",");
            }
            token = op;
        }
        // Check for colon syntax: op:arg
        char *colon2 = strchr(token, ':');
        char opName[32], opArg[32];
        if (colon2) {
            int len = colon2 - token;
            if (len >= 32) len = 31;
            strncpy(opName, token, len);
            opName[len] = '\0';
            strcpy(opArg, colon2 + 1);
        } else {
            strcpy(opName, token);
            opArg[0] = '\0';
        }
        if (strcmp(opName, "push") == 0) {
            if (opArg[0] == '\0')
                printf("push requires an argument\n");
            else
                intStackPerspectivePush(stackPerspective, atoi(opArg));
        } else if (strcmp(opName, "pop") == 0) {
            int ok;
            int val = intStackPerspectivePop(stackPerspective, &ok);
            if (ok)
                printf("Popped: %d\n", val);
            else
                printf("Stack is empty\n");
        } else if (strcmp(opName, "add") == 0) {
            if (!intStackAdd(stackPerspective->physical))
                printf("Addition failed\n");
        } else if (strcmp(opName, "sub") == 0) {
            if (!intStackSub(stackPerspective->physical))
                printf("Subtraction failed\n");
        } else if (strcmp(opName, "mul") == 0) {
            if (!intStackMul(stackPerspective->physical))
                printf("Multiplication failed\n");
        } else if (strcmp(opName, "div") == 0) {
            if (!intStackDiv(stackPerspective->physical))
                printf("Division failed\n");
        } else if (strcmp(opName, "print") == 0) {
            printf("@%s: ", selector);
            intStackPerspectivePrint(stackPerspective);
        } else if (strcmp(opName, "lifo") == 0) {
            intStackPerspectiveSetPerspective(stackPerspective, "lifo");
            printf("@%s perspective set: LIFO\n", selector);
        } else if (strcmp(opName, "fifo") == 0) {
            intStackPerspectiveSetPerspective(stackPerspective, "fifo");
            printf("@%s perspective set: FIFO\n", selector);
        } else if (strcmp(opName, "flip") == 0) {
            intStackPerspectiveFlip(stackPerspective);
            {
                char modeUpper[20];
                toUppercase(stackPerspective->perspective, modeUpper, sizeof(modeUpper));
                printf("@%s perspective flipped to: %s\n", selector, modeUpper);
            }
        } else {
            printf("Unknown compound command: %s\n", opName);
        }
        token = strtok(NULL, " ");
    }
}

void executeSpawnCommand(const char *cmd) {
    // Execute a spawn command.
    char *copy = strdup(cmd);
    char *token = strtok(copy, " ");
    if (!token) { 
        free(copy);
        return;
    }
    if (strcmp(token, "list") == 0) {
        printf("[spawn] list command executed.\n");
    } else if (strcmp(token, "add") == 0) {
        token = strtok(NULL, " ");
        if (token)
            printf("[spawn] add command: %s\n", token);
    } else if (strcmp(token, "pause") == 0) {
        token = strtok(NULL, " ");
        if (token)
            printf("[spawn] pause command: %s\n", token);
    } else if (strcmp(token, "resume") == 0) {
        token = strtok(NULL, " ");
        if (token)
            printf("[spawn] resume command: %s\n", token);
    } else if (strcmp(token, "stop") == 0) {
        token = strtok(NULL, " ");
        if (token)
            printf("[spawn] stop command: %s\n", token);
    } else {
        printf("[spawn] Unknown command: %s\n", token);
    }
    free(copy);
}
