#include "stacks/stacks.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Global memory for STORE/LOAD operations (for int stacks)
int globalMemory[1024] = {0};

// ---------- IntStack Implementation ----------

IntStack *newIntStack() {
    IntStack *s = malloc(sizeof(IntStack));
    if (!s) { perror("malloc newIntStack"); exit(EXIT_FAILURE); }
    s->capacity = 16;
    s->size = 0;
    s->data = malloc(sizeof(int) * s->capacity);
    if (!s->data) { perror("malloc newIntStack->data"); exit(EXIT_FAILURE); }
    strcpy(s->mode, "lifo");
    return s;
}

void intStackPush(IntStack *s, int val) {
    if (s->size >= s->capacity) {
        s->capacity *= 2;
        int *newData = realloc(s->data, sizeof(int) * s->capacity);
        if (!newData) { perror("realloc failed in intStackPush"); exit(EXIT_FAILURE); }
        s->data = newData;
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
    for (int i = 0; i < s->size; i++) {
        printf("%d ", s->data[i]);
    }
    printf("\n");
}

void intStackSetMode(IntStack *s, const char *mode) {
    if (strcmp(mode, "lifo") == 0 || strcmp(mode, "fifo") == 0)
        strcpy(s->mode, mode);
}

void intStackFlip(IntStack *s) {
    // Reverse the underlying array.
    for (int i = 0, j = s->size - 1; i < j; i++, j--) {
        int tmp = s->data[i];
        s->data[i] = s->data[j];
        s->data[j] = tmp;
    }
    // Toggle mode flag: if it was "lifo", change it to "fifo", and vice versa.
    if (strcmp(s->mode, "lifo") == 0)
        strcpy(s->mode, "fifo");
    else
        strcpy(s->mode, "lifo");
}

// --- Forth-like Operations for IntStack ---

int intStackAdd(IntStack *s) {
    int ok1, ok2;
    int b = intStackPop(s, &ok1);
    int a = intStackPop(s, &ok2);
    if (!ok1 || !ok2) return 0;
    intStackPush(s, a + b);
    return 1;
}

int intStackSub(IntStack *s) {
    int ok1, ok2;
    int b = intStackPop(s, &ok1);
    int a = intStackPop(s, &ok2);
    if (!ok1 || !ok2) return 0;
    intStackPush(s, a - b);
    return 1;
}

int intStackMul(IntStack *s) {
    int ok1, ok2;
    int b = intStackPop(s, &ok1);
    int a = intStackPop(s, &ok2);
    if (!ok1 || !ok2) return 0;
    intStackPush(s, a * b);
    return 1;
}

int intStackDiv(IntStack *s) {
    int ok1, ok2;
    int b = intStackPop(s, &ok1);
    if (!ok1 || b == 0) {
        printf("Division error\n");
        return 0;
    }
    int a = intStackPop(s, &ok2);
    if (!ok2) return 0;
    intStackPush(s, a / b);
    return 1;
}

int intStackTuck(IntStack *s) {
    if (s->size < 2) return 0;
    int top = s->data[s->size - 1];
    if (s->size >= s->capacity) {
        s->capacity *= 2;
        int *newData = realloc(s->data, sizeof(int) * s->capacity);
        if (!newData) { perror("realloc in tuck"); exit(EXIT_FAILURE); }
        s->data = newData;
    }
    memmove(&s->data[s->size], &s->data[s->size - 1], sizeof(int));
    s->data[s->size - 1] = top;
    s->size++;
    return 1;
}

int intStackPick(IntStack *s, int n) {
    if (n < 0 || n >= s->size) return 0;
    intStackPush(s, s->data[s->size - 1 - n]);
    return 1;
}

int intStackRoll(IntStack *s, int n) {
    if (n < 0 || n >= s->size) return 0;
    int idx = s->size - 1 - n;
    int val = s->data[idx];
    memmove(&s->data[idx], &s->data[idx+1], sizeof(int) * (s->size - idx - 1));
    s->data[s->size - 1] = val;
    return 1;
}

int intStackOver2(IntStack *s) {
    if (s->size < 4) return 0;
    intStackPush(s, s->data[s->size - 4]);
    intStackPush(s, s->data[s->size - 4]);
    return 1;
}

int intStackDrop2(IntStack *s) {
    if (s->size < 2) return 0;
    s->size -= 2;
    return 1;
}

int intStackSwap2(IntStack *s) {
    if (s->size < 4) return 0;
    int i = s->size - 4;
    int tmp = s->data[i];
    s->data[i] = s->data[i+2];
    s->data[i+2] = tmp;
    tmp = s->data[i+1];
    s->data[i+1] = s->data[i+3];
    s->data[i+3] = tmp;
    return 1;
}

int intStackDepthValue(IntStack *s) {
    return s->size;
}

// --- Memory Operations for IntStack ---
int intStackStore(IntStack *s) {
    if (intStackDepth(s) < 2) return 0;
    int ok;
    int address = intStackPop(s, &ok);
    if (!ok) return 0;
    int value = intStackPop(s, &ok);
    if (!ok) return 0;
    if (address < 0 || address >= 1024) {
        printf("Address %d out of bounds\n", address);
        return 0;
    }
    globalMemory[address] = value;
    return 1;
}

int intStackLoad(IntStack *s) {
    if (intStackDepth(s) < 1) return 0;
    int ok;
    int address = intStackPop(s, &ok);
    if (!ok) return 0;
    if (address < 0 || address >= 1024) {
        printf("Address %d out of bounds\n", address);
        return 0;
    }
    int value = globalMemory[address];
    intStackPush(s, value);
    return 1;
}

// --- Bitwise Operations for IntStack ---
int intStackAnd(IntStack *s) {
    if (intStackDepth(s) < 2) return 0;
    int ok1, ok2;
    int b = intStackPop(s, &ok1);
    int a = intStackPop(s, &ok2);
    if (!ok1 || !ok2) return 0;
    intStackPush(s, a & b);
    return 1;
}

int intStackOr(IntStack *s) {
    if (intStackDepth(s) < 2) return 0;
    int ok1, ok2;
    int b = intStackPop(s, &ok1);
    int a = intStackPop(s, &ok2);
    if (!ok1 || !ok2) return 0;
    intStackPush(s, a | b);
    return 1;
}

int intStackXor(IntStack *s) {
    if (intStackDepth(s) < 2) return 0;
    int ok1, ok2;
    int b = intStackPop(s, &ok1);
    int a = intStackPop(s, &ok2);
    if (!ok1 || !ok2) return 0;
    intStackPush(s, a ^ b);
    return 1;
}

int intStackShl(IntStack *s) {
    if (intStackDepth(s) < 2) return 0;
    int ok1, ok2;
    int bits = intStackPop(s, &ok1);
    int value = intStackPop(s, &ok2);
    if (!ok1 || !ok2) return 0;
    intStackPush(s, value << bits);
    return 1;
}

int intStackShr(IntStack *s) {
    if (intStackDepth(s) < 2) return 0;
    int ok1, ok2;
    int bits = intStackPop(s, &ok1);
    int value = intStackPop(s, &ok2);
    if (!ok1 || !ok2) return 0;
    intStackPush(s, value >> bits);
    return 1;
}

// ---------- StringStack Implementation ----------

StringStack *newStringStack() {
    StringStack *s = malloc(sizeof(StringStack));
    if (!s) { perror("malloc newStringStack"); exit(EXIT_FAILURE); }
    s->capacity = 16;
    s->size = 0;
    s->data = malloc(sizeof(char*) * s->capacity);
    if (!s->data) { perror("malloc newStringStack->data"); exit(EXIT_FAILURE); }
    strcpy(s->mode, "lifo");
    return s;
}

void stringStackPush(StringStack *s, const char *val) {
    if (s->size >= s->capacity) {
        s->capacity *= 2;
        char **newData = realloc(s->data, sizeof(char*) * s->capacity);
        if (!newData) { perror("realloc in stringStackPush"); exit(EXIT_FAILURE); }
        s->data = newData;
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
        memmove(s->data, s->data + 1, sizeof(char*) * (s->size - 1));
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
    for (int i = 0; i < s->size; i++) {
        printf("  %s\n", s->data[i]);
    }
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

// --- Forth-like Operations for StringStack ---

int stringStackAdd(StringStack *s) {
    if (s->size < 2) return 0;
    int success;
    char *b = stringStackPop(s, &success);
    if (!success) return 0;
    char *a = stringStackPop(s, &success);
    if (!success) {
        free(b);
        return 0;
    }
    size_t len = strlen(a) + strlen(b) + 1;
    char *result = malloc(len);
    snprintf(result, len, "%s%s", a, b);
    stringStackPush(s, result);
    free(a);
    free(b);
    free(result);
    return 1;
}

int stringStackSub(StringStack *s, const char *trimChar) {
    if (s->size < 1) return 0;
    int success;
    char *top = stringStackPop(s, &success);
    if (!success) return 0;
    size_t len = strlen(top);
    while (len > 0 && strcmp(&top[len - 1], trimChar) == 0) {
        top[len - 1] = '\0';
        len--;
    }
    stringStackPush(s, top);
    free(top);
    return 1;
}

int stringStackMul(StringStack *s, int n) {
    if (s->size < 1 || n < 0) return 0;
    int success;
    char *str = stringStackPop(s, &success);
    if (!success) return 0;
    size_t len = strlen(str);
    size_t total = len * n + 1;
    char *result = malloc(total);
    result[0] = '\0';
    for (int i = 0; i < n; i++) {
        strcat(result, str);
    }
    stringStackPush(s, result);
    free(str);
    free(result);
    return 1;
}

int stringStackDiv(StringStack *s, const char *delim) {
    if (s->size < 1) return 0;
    int success;
    char *str = stringStackPop(s, &success);
    if (!success) return 0;
    char *copy = strdup(str);
    char *token = strtok(copy, delim);
    char *result = NULL;
    size_t resLen = 0;
    while (token) {
        size_t tokenLen = strlen(token);
        result = realloc(result, resLen + tokenLen + 2);
        if (resLen == 0)
            strcpy(result, token);
        else {
            result[resLen] = ' ';
            strcpy(result + resLen + 1, token);
        }
        resLen = strlen(result);
        token = strtok(NULL, delim);
    }
    free(copy);
    stringStackPush(s, result ? result : "");
    free(str);
    free(result);
    return 1;
}

// --- Cleanup Functions ---
void freeIntStack(IntStack *s) {
    if (s) {
        free(s->data);
        free(s);
    }
}

void freeStringStack(StringStack *s) {
    if (s) {
        for (int i = 0; i < s->size; i++) {
            free(s->data[i]);
        }
        free(s->data);
        free(s);
    }
}

// ---------- New: IntStackPerspective Implementation ----------

IntStackPerspective *newIntStackPerspective(IntStack *physical) {
    IntStackPerspective *sp = malloc(sizeof(IntStackPerspective));
    if (!sp) {
        perror("malloc newIntStackPerspective");
        exit(EXIT_FAILURE);
    }
    sp->physical = physical;
    sp->startIndex = 0;
    strcpy(sp->perspective, "lifo");  // default perspective is lifo
    return sp;
}

void freeIntStackPerspective(IntStackPerspective *sp) {
    if (sp) {
        // Note: Do not free the underlying physical stack.
        free(sp);
    }
}

void intStackPerspectivePush(IntStackPerspective *sp, int val) {
    // Always append to the underlying physical stack.
    intStackPush(sp->physical, val);
}

int intStackPerspectivePop(IntStackPerspective *sp, int *success) {
    if (strcmp(sp->perspective, "lifo") == 0) {
        return intStackPop(sp->physical, success);
    } else if (strcmp(sp->perspective, "fifo") == 0) {
        int available = sp->physical->size - sp->startIndex;
        if (available <= 0) {
            *success = 0;
            return 0;
        }
        int val = sp->physical->data[sp->startIndex];
        sp->startIndex++;
        *success = 1;
        // Reset the view when exhausted.
        if (sp->startIndex == sp->physical->size) {
            sp->startIndex = 0;
            sp->physical->size = 0;
        }
        return val;
    } else {
        *success = 0;
        return 0;
    }
}

void intStackPerspectiveSetPerspective(IntStackPerspective *sp, const char *persp) {
    if (strcmp(persp, "lifo") == 0 || strcmp(persp, "fifo") == 0) {
        // When switching from FIFO to LIFO, compact the stack.
        if (strcmp(sp->perspective, "fifo") == 0 && strcmp(persp, "lifo") == 0) {
            int newSize = sp->physical->size - sp->startIndex;
            memmove(sp->physical->data, sp->physical->data + sp->startIndex, sizeof(int) * newSize);
            sp->physical->size = newSize;
            sp->startIndex = 0;
        }
        strcpy(sp->perspective, persp);
    } else {
        fprintf(stderr, "Unsupported perspective: %s\n", persp);
    }
}

void intStackPerspectiveFlip(IntStackPerspective *sp) {
    if (strcmp(sp->perspective, "lifo") == 0) {
        strcpy(sp->perspective, "fifo");
        sp->startIndex = 0;
    } else if (strcmp(sp->perspective, "fifo") == 0) {
        int newSize = sp->physical->size - sp->startIndex;
        memmove(sp->physical->data, sp->physical->data + sp->startIndex, sizeof(int) * newSize);
        sp->physical->size = newSize;
        sp->startIndex = 0;
        strcpy(sp->perspective, "lifo");
    }
}

void intStackPerspectivePrint(IntStackPerspective *sp) {
    // Display the default name "@dstack:" for consistency.
    printf("@dstack: ");
    IntStack *s = sp->physical;
    if (strcmp(sp->perspective, "fifo") == 0) {
        if (s->size - sp->startIndex > 0) {
            // In FIFO mode, bracket the element at startIndex.
            printf("[ %d ] ", s->data[sp->startIndex]);
            for (int i = sp->startIndex + 1; i < s->size; i++) {
                printf("%d ", s->data[i]);
            }
        } else {
            printf("Empty ");
        }
    } else if (strcmp(sp->perspective, "lifo") == 0) {
        if (s->size > 0) {
            for (int i = 0; i < s->size - 1; i++) {
                printf("%d ", s->data[i]);
            }
            // In LIFO mode, bracket the last element.
            printf("[ %d ]", s->data[s->size - 1]);
        } else {
            printf("Empty");
        }
    } else {
        // Fallback: print all elements normally.
        for (int i = 0; i < s->size; i++) {
            printf("%d ", s->data[i]);
        }
    }
    printf("\n");
}
