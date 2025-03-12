#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
#include <ctype.h>
#include <unistd.h>

// ------------------------------
// Dynamic Int Stack Implementation
// ------------------------------
typedef struct {
    int *data;
    int size;
    int capacity;
    char mode[5]; // "lifo" or "fifo"
} IntStack;

IntStack *newIntStack() {
    IntStack *s = malloc(sizeof(IntStack));
    s->capacity = 16;
    s->size = 0;
    s->data = malloc(sizeof(int) * s->capacity);
    strcpy(s->mode, "lifo");
    return s;
}

void intStackPush(IntStack *s, int val) {
    if (s->size >= s->capacity) {
        s->capacity *= 2;
        s->data = realloc(s->data, sizeof(int) * s->capacity);
    }
    s->data[s->size++] = val;
}

int intStackPop(IntStack *s, int *success) {
    if (s->size == 0) {
        *success = 0;
        return 0;
    }
    if (strcmp(s->mode, "fifo") == 0) {
        int val = s->data[0];
        memmove(s->data, s->data + 1, sizeof(int) * (s->size - 1));
        s->size--;
        *success = 1;
        return val;
    }
    *success = 1;
    return s->data[--s->size];
}

int intStackDepth(IntStack *s) {
    return s->size;
}

void intStackPrint(IntStack *s) {
    printf("IntStack (%s mode): ", s->mode);
    for (int i = 0; i < s->size; i++)
        printf("%d ", s->data[i]);
    printf("\n");
}

void intStackSetMode(IntStack *s, const char *mode) {
    if (strcmp(mode, "lifo") == 0 || strcmp(mode, "fifo") == 0) {
        strcpy(s->mode, mode);
    }
}

void intStackFlip(IntStack *s) {
    for (int i = 0, j = s->size - 1; i < j; i++, j--) {
        int tmp = s->data[i];
        s->data[i] = s->data[j];
        s->data[j] = tmp;
    }
}

// Additional Forth-like operations (e.g., tuck, pick, roll) can be added similarly.
// For brevity, we implement only a few:
int intStackAdd(IntStack *s) {
    int ok1, ok2;
    int b = intStackPop(s, &ok1);
    int a = intStackPop(s, &ok2);
    if (!ok1 || !ok2) return 0;
    intStackPush(s, a + b);
    return 1;
}

// ------------------------------
// Dynamic String Stack Implementation
// ------------------------------
typedef struct {
    char **data;
    int size;
    int capacity;
    char mode[5]; // "lifo" or "fifo"
} StringStack;

StringStack *newStringStack() {
    StringStack *s = malloc(sizeof(StringStack));
    s->capacity = 16;
    s->size = 0;
    s->data = malloc(sizeof(char*) * s->capacity);
    strcpy(s->mode, "lifo");
    return s;
}

void stringStackPush(StringStack *s, const char *val) {
    if (s->size >= s->capacity) {
        s->capacity *= 2;
        s->data = realloc(s->data, sizeof(char*) * s->capacity);
    }
    s->data[s->size++] = strdup(val);
}

char *stringStackPop(StringStack *s, int *success) {
    if (s->size == 0) {
        *success = 0;
        return NULL;
    }
    char *val;
    if (strcmp(s->mode, "fifo") == 0) {
        val = s->data[0];
        for (int i = 1; i < s->size; i++) {
            s->data[i-1] = s->data[i];
        }
        s->size--;
        *success = 1;
        return val;
    }
    *success = 1;
    return s->data[--s->size];
}

int stringStackDepth(StringStack *s) {
    return s->size;
}

void stringStackPrint(StringStack *s) {
    printf("StringStack (%s mode):\n", s->mode);
    for (int i = 0; i < s->size; i++)
        printf("%s\n", s->data[i]);
}

void stringStackSetMode(StringStack *s, const char *mode) {
    if (strcmp(mode, "lifo") == 0 || strcmp(mode, "fifo") == 0)
        strcpy(s->mode, mode);
}

void stringStackFlip(StringStack *s) {
    for (int i = 0, j = s->size - 1; i < j; i++, j--) {
        char *tmp = s->data[i];
        s->data[i] = s->data[j];
        s->data[j] = tmp;
    }
}

// ------------------------------
// Spawn (Goroutine) Implementation using pthreads
// ------------------------------
typedef struct {
    char *name;
    pthread_t thread;
    pthread_mutex_t mtx;
    pthread_cond_t cond;
    int paused;
    int stop;
    char *script; // container for a script
} Spawn;

Spawn *newSpawn(const char *name);

void *spawnMain(void *arg);
void spawnRunScript(Spawn *sp);

Spawn *newSpawn(const char *name) {
    Spawn *sp = malloc(sizeof(Spawn));
    sp->name = strdup(name);
    sp->paused = 0;
    sp->stop = 0;
    sp->script = NULL;
    pthread_mutex_init(&sp->mtx, NULL);
    pthread_cond_init(&sp->cond, NULL);
    pthread_create(&sp->thread, NULL, spawnMain, sp);
    return sp;
}

void *spawnMain(void *arg) {
    Spawn *sp = arg;
    while (1) {
        pthread_mutex_lock(&sp->mtx);
        while (sp->paused && !sp->stop) {
            pthread_cond_wait(&sp->cond, &sp->mtx);
        }
        if (sp->stop) {
            pthread_mutex_unlock(&sp->mtx);
            break;
        }
        // Check if there is a script to run
        if (sp->script) {
            printf("[spawn:%s] Executing script:\n%s\n", sp->name, sp->script);
            // For simplicity, split script by newline and print each line.
            char *scriptCopy = strdup(sp->script);
            char *line = strtok(scriptCopy, "\n");
            while (line) {
                printf("[spawn:%s] > %s\n", sp->name, line);
                // In a more complete implementation, we would pass this line to the interpreter.
                line = strtok(NULL, "\n");
            }
            free(scriptCopy);
            free(sp->script);
            sp->script = NULL;
        }
        pthread_mutex_unlock(&sp->mtx);
        sleep(1);
    }
    printf("[spawn:%s] Exiting.\n", sp->name);
    return NULL;
}

void spawnPause(Spawn *sp) {
    pthread_mutex_lock(&sp->mtx);
    sp->paused = 1;
    pthread_mutex_unlock(&sp->mtx);
}

void spawnResume(Spawn *sp) {
    pthread_mutex_lock(&sp->mtx);
    sp->paused = 0;
    pthread_cond_signal(&sp->cond);
    pthread_mutex_unlock(&sp->mtx);
}

void spawnStop(Spawn *sp) {
    pthread_mutex_lock(&sp->mtx);
    sp->stop = 1;
    pthread_cond_signal(&sp->cond);
    pthread_mutex_unlock(&sp->mtx);
}

void spawnRunScript(Spawn *sp) {
    pthread_mutex_lock(&sp->mtx);
    pthread_cond_signal(&sp->cond);
    pthread_mutex_unlock(&sp->mtx);
}

// ------------------------------
// Global Spawn Manager (Simple linked list)
// ------------------------------
typedef struct SpawnNode {
    Spawn *sp;
    struct SpawnNode *next;
} SpawnNode;

SpawnNode *spawnList = NULL;

void spawnManagerAdd(Spawn *sp) {
    SpawnNode *node = malloc(sizeof(SpawnNode));
    node->sp = sp;
    node->next = spawnList;
    spawnList = node;
    printf("Added spawn '%s'\n", sp->name);
}

Spawn *spawnManagerFind(const char *name) {
    SpawnNode *cur = spawnList;
    while (cur) {
        if (strcmp(cur->sp->name, name) == 0) {
            return cur->sp;
        }
        cur = cur->next;
    }
    return NULL;
}

void spawnManagerList() {
    printf("Spawns:\n");
    SpawnNode *cur = spawnList;
    while (cur) {
        printf("  %s\n", cur->sp->name);
        cur = cur->next;
    }
}

// ------------------------------
// Simple Command Interpreter (for iual)
// ------------------------------
typedef struct {
    char *name;
    char *type; // "int", "str", "spawn"
} StackSelector;

void executeSpawnCommand(const char *cmd); // forward declaration

// Global pointers for default stacks and spawn manager
// (We use simple arrays for int and string stacks.)
IntStack *dstack; // data stack (default int stack)
IntStack *rstack; // return stack (default int stack)

// For simplicity, we maintain string stacks in a global hash (here using a simple array)
#define MAX_STR_STACKS 16
typedef struct {
    char *name;
    StringStack *stack;
} NamedStringStack;
NamedStringStack strStacksArr[MAX_STR_STACKS];
int strStacksCount = 0;

// Helper to add a new string stack
void addStringStack(const char *name) {
    if (strStacksCount < MAX_STR_STACKS) {
        strStacksArr[strStacksCount].name = strdup(name);
        strStacksArr[strStacksCount].stack = newStringStack();
        strStacksCount++;
        printf("Created new string stack '%s'\n", name);
    }
}

// Helper to find a string stack by name
StringStack *findStringStack(const char *name) {
    for (int i = 0; i < strStacksCount; i++) {
        if (strcmp(strStacksArr[i].name, name) == 0) {
            return strStacksArr[i].stack;
        }
    }
    return NULL;
}

// Compound command parser: if input starts with '@' and contains ':'
void processCompoundCommand(char *input) {
    // Example: "@dstack: push:1 pop add"
    char *colon = strchr(input, ':');
    if (!colon) return;
    *colon = '\0';
    char *selector = input + 1; // skip '@'
    char *ops = colon + 1;
    // For simplicity, we only implement compound commands for int stacks and spawn.
    if (strcmp(selector, "spawn") == 0) {
        // Process spawn compound commands.
        // Tokenize ops.
        char *token = strtok(ops, " ");
        while (token) {
            // For function-like syntax: e.g., div(10,2)
            if (strchr(token, '(') && token[strlen(token)-1] == ')') {
                char *paren = strchr(token, '(');
                *paren = '\0';
                char *op = token;
                char *argList = paren + 1;
                argList[strlen(argList)-1] = '\0'; // remove ')'
                // For spawn, we assume op is "bring" with a script.
                if (strcmp(op, "bring") == 0) {
                    // Expect arguments like: str,@sstack
                    char *arg = argList;
                    char *comma = strchr(arg, ',');
                    if (comma) {
                        *comma = '\0';
                        char *srcType = arg;
                        char *srcStackName = comma + 1;
                        if (srcStackName[0] == '@') {
                            srcStackName++;
                        }
                        if (strcmp(srcType, "str") != 0) {
                            printf("For spawn bring, only string scripts are supported.\n");
                        } else {
                            StringStack *ss = findStringStack(srcStackName);
                            if (!ss) {
                                printf("No string stack named '%s'\n", srcStackName);
                            } else {
                                // Pop all lines (simulate script extraction)
                                char *script = NULL;
                                int first = 1;
                                int success;
                                while (1) {
                                    char *line = stringStackPop(ss, &success);
                                    if (!success) break;
                                    // Prepend the line (we reverse order)
                                    if (first) {
                                        script = strdup(line);
                                        first = 0;
                                    } else {
                                        char *temp = script;
                                        script = malloc(strlen(line) + strlen(temp) + 2);
                                        sprintf(script, "%s\n%s", line, temp);
                                        free(temp);
                                    }
                                    free(line);
                                }
                                // Store script in spawn "spawn"
                                Spawn *sp = spawnManagerFind("spawn");
                                if (sp) {
                                    free(sp->script);
                                    sp->script = script;
                                    printf("Script stored in spawn 'spawn'. Use run to execute.\n");
                                }
                            }
                        }
                    }
                }
            } else if (strcmp(token, "run") == 0) {
                // For spawn, run the script.
                Spawn *sp = spawnManagerFind("spawn");
                if (sp) {
                    spawnRunScript(sp);
                }
            } else {
                // For int stack compound commands.
                // Token may be in the form op:arg (e.g., push:1)
                char *colon2 = strchr(token, ':');
                char op[32], arg[32];
                if (colon2) {
                    strncpy(op, token, colon2 - token);
                    op[colon2 - token] = '\0';
                    strcpy(arg, colon2 + 1);
                } else {
                    strcpy(op, token);
                    arg[0] = '\0';
                }
                IntStack *stack = dstack; // assume dstack selected for this demo
                if (strcmp(op, "push") == 0) {
                    if (arg[0] == '\0') {
                        printf("push requires an argument\n");
                    } else {
                        int val = atoi(arg);
                        intStackPush(stack, val);
                    }
                } else if (strcmp(op, "pop") == 0) {
                    int ok;
                    int val = intStackPop(stack, &ok);
                    if (ok) {
                        printf("Popped: %d\n", val);
                    } else {
                        printf("Stack empty\n");
                    }
                } else if (strcmp(op, "add") == 0) {
                    if (!intStackAdd(stack)) {
                        printf("Addition failed\n");
                    }
                } else {
                    printf("Unknown compound op: %s\n", op);
                }
            }
            token = strtok(NULL, " ");
		}
		return;
	}
	// For this simplified example, if selector is not "spawn", we assume it's "dstack" or "rstack" (int stacks).
	IntStack *selStack = NULL;
	if (strcmp(selector, "dstack") == 0) {
		selStack = dstack;
	} else if (strcmp(selector, "rstack") == 0) {
		selStack = rstack;
	} else {
		printf("Compound commands for selector '%s' not implemented.\n", selector);
		return;
	}
	// Process tokens for int stacks:
	char *token = strtok(ops, " ");
	while (token) {
		// Function-like syntax?
		if (strchr(token, '(') && token[strlen(token)-1] == ')') {
			char *paren = strchr(token, '(');
			*paren = '\0';
			char *funcOp = token;
			char *argList = paren + 1;
			argList[strlen(argList)-1] = '\0';
			char *argToken = strtok(argList, ",");
			while (argToken) {
				intStackPush(selStack, atoi(argToken));
				argToken = strtok(NULL, ",");
			}
			strcpy(token, funcOp);
		}
		// Check for colon syntax: op:arg
		char *colon = strchr(token, ':');
		char op[32], arg[32];
		if (colon) {
			strncpy(op, token, colon - token);
			op[colon - token] = '\0';
			strcpy(arg, colon + 1);
		} else {
			strcpy(op, token);
			arg[0] = '\0';
		}
		if (strcmp(op, "push") == 0) {
			if (arg[0] == '\0') {
				printf("push requires an argument\n");
			} else {
				intStackPush(selStack, atoi(arg));
			}
		} else if (strcmp(op, "pop") == 0) {
			int ok;
			int val = intStackPop(selStack, &ok);
			if (ok)
				printf("Popped: %d\n", val);
			else
				printf("Stack empty\n");
		} else if (strcmp(op, "add") == 0) {
			if (!intStackAdd(selStack)) {
				printf("Addition failed\n");
			}
		} else if (strcmp(op, "print") == 0) {
			intStackPrint(selStack);
		} else {
			printf("Unknown compound op: %s\n", op);
		}
		token = strtok(NULL, " ");
	}
}

// ------------------------------
// Main Interactive Loop
// ------------------------------
int main(void) {
	// Print header
	printf("iual v0.0.1\n");
	printf("iual is an exceedingly trivial interactive ual 0.0.1 interpreter\n");

	// Create default int stacks.
	dstack = newIntStack();
	rstack = newIntStack();

	// Create a default string stack for scripts.
	addStringStack("sstack");

	// Initialize spawn manager and create a default spawn called "spawn".
	Spawn *defaultSpawn = newSpawn("spawn");
	spawnManagerAdd(defaultSpawn);

	// Set up global string stacks pointer
	globalStrStacks = malloc(sizeof(StringStack*) * MAX_STR_STACKS);
	// (For simplicity, we assume strStacksArr is our global store.)

	// Main command loop.
	char line[256];
	while (1) {
		printf("> ");
		if (!fgets(line, sizeof(line), stdin))
			break;
		// Remove trailing newline.
		line[strcspn(line, "\n")] = 0;
		if (strlen(line) == 0)
			continue;
		// If compound command: starts with '@' and contains ':'
		if (line[0] == '@' && strchr(line, ':')) {
			processCompoundCommand(line);
			continue;
		}
		// Otherwise, parse global commands (very simplified).
		char *cmd = strtok(line, " ");
		if (!cmd) continue;
		if (strcmp(cmd, "new") == 0) {
			char *name = strtok(NULL, " ");
			char *type = strtok(NULL, " ");
			if (!name || !type) {
				printf("Usage: new <stack name> <int|str|float>\n");
				continue;
			}
			if (strcmp(type, "int") == 0) {
				dstack = newIntStack(); // For demo, we replace dstack
				printf("Created new int stack '%s'\n", name);
			} else if (strcmp(type, "str") == 0) {
				addStringStack(name);
			} else {
				printf("Only int and str stacks implemented in this demo.\n");
			}
		} else if (strcmp(cmd, "spawn") == 0) {
			char *name = strtok(NULL, " ");
			if (!name) {
				printf("Usage: spawn <goroutine name>\n");
				continue;
			}
			Spawn *sp = newSpawn(name);
			spawnManagerAdd(sp);
		} else if (strcmp(cmd, "list") == 0) {
			spawnManagerList();
		} else if (strcmp(cmd, "quit") == 0) {
			break;
		} else {
			printf("Unknown global command: %s\n", cmd);
		}
	}
	return 0;
}
