#include "stacks.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* ---------- IntStack Implementation ---------- */

IntStack *newIntStack(void) {
    IntStack *s = malloc(sizeof(IntStack));
    if (!s) { 
        perror("malloc newIntStack"); 
        exit(EXIT_FAILURE); 
    }
    s->capacity = 16;
    s->size = 0;
    s->data = malloc(sizeof(int) * s->capacity);
    if (!s->data) { 
        perror("malloc newIntStack->data"); 
        exit(EXIT_FAILURE); 
    }
    strcpy(s->mode, "lifo");
    return s;
}

/*
   Modified intStackPush:
   - In FIFO mode, new values are inserted at index 0.
   - In LIFO mode, new values are appended.
*/
void intStackPush(IntStack *s, int val) {
    if (strcmp(s->mode, "fifo") == 0) {
        if (s->size >= s->capacity) {
            s->capacity *= 2;
            int *newData = realloc(s->data, sizeof(int) * s->capacity);
            if (!newData) { 
                perror("realloc failed in intStackPush (fifo)"); 
                exit(EXIT_FAILURE); 
            }
            s->data = newData;
        }
        // Shift current elements one position to the right.
        memmove(s->data + 1, s->data, sizeof(int) * s->size);
        s->data[0] = val;
        s->size++;
    } else {  // lifo mode
        if (s->size >= s->capacity) {
            s->capacity *= 2;
            int *newData = realloc(s->data, sizeof(int) * s->capacity);
            if (!newData) { 
                perror("realloc failed in intStackPush (lifo)"); 
                exit(EXIT_FAILURE); 
            }
            s->data = newData;
        }
        s->data[s->size++] = val;
    }
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
    /* This function is no longer used for perspective flipping.
       It used to reverse the physical order, but now we want flip to act
       solely as a toggle of the current perspective.
    */
    /* (Empty implementation) */
}

void freeIntStack(IntStack *s) {
    if (s) {
        free(s->data);
        free(s);
    }
}

/* ---------- Forth-like Operations for IntStack ---------- */

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
        if (!newData) { 
            perror("realloc in tuck"); 
            exit(EXIT_FAILURE); 
        }
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
    memmove(&s->data[idx], &s->data[idx + 1], sizeof(int) * (s->size - idx - 1));
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

/* ---------- Memory Operations for IntStack ---------- */

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

/* ---------- Bitwise Operations for IntStack ---------- */

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

/* ---------- IntStackPerspective Implementation ---------- */

IntStackPerspective *newIntStackPerspective(IntStack *s, const char *name) {
    IntStackPerspective *persp = malloc(sizeof(IntStackPerspective));
    if (!persp) { 
        perror("malloc newIntStackPerspective"); 
        exit(EXIT_FAILURE); 
    }
    persp->physical = s;
    persp->startIndex = 0;
    strcpy(persp->perspective, s->mode);
    strncpy(persp->name, name, sizeof(persp->name) - 1);
    persp->name[sizeof(persp->name) - 1] = '\0';

    // Register in the global list.
    extern IntStackNode *globalIntStackList;
    IntStackNode *node = malloc(sizeof(IntStackNode));
    if (!node) { perror("malloc IntStackNode"); exit(EXIT_FAILURE); }
    node->name = strdup(name);
    node->persp = persp;
    node->next = globalIntStackList;
    globalIntStackList = node;

    return persp;
}

void intStackPerspectivePush(IntStackPerspective *sp, int val) {
    intStackPush(sp->physical, val);
}

int intStackPerspectivePop(IntStackPerspective *sp, int *success) {
    return intStackPop(sp->physical, success);
}

void intStackPerspectivePrint(IntStackPerspective *sp) {
    printf("@%s: ", sp->name);
    if (sp->physical->size == 0) {
        printf("Empty");
    } else {
        if (strcmp(sp->perspective, "fifo") == 0) {
            if (sp->physical->size > sp->startIndex) {
                printf("[ %d ] ", sp->physical->data[sp->startIndex]);
                for (int i = sp->startIndex + 1; i < sp->physical->size; i++) {
                    printf("%d ", sp->physical->data[i]);
                }
            }
        } else { // lifo
            if (sp->physical->size > 0) {
                for (int i = 0; i < sp->physical->size - 1; i++) {
                    printf("%d ", sp->physical->data[i]);
                }
                printf("[ %d ]", sp->physical->data[sp->physical->size - 1]);
            }
        }
    }
    printf("\n");
}

/*
   Updated intStackPerspectiveSetPerspective:
   - When setting FIFO mode, reset startIndex to 0.
   - When setting LIFO mode, leave startIndex unchanged.
*/
void intStackPerspectiveSetPerspective(IntStackPerspective *sp, const char *mode) {
    if (strcmp(mode, "fifo") == 0) {
        strcpy(sp->perspective, "fifo");
        intStackSetMode(sp->physical, "fifo");
        sp->startIndex = 0;
    } else if (strcmp(mode, "lifo") == 0) {
        strcpy(sp->perspective, "lifo");
        intStackSetMode(sp->physical, "lifo");
        /* In LIFO mode, the physical order is preserved */
    }
}

/*
   Updated intStackPerspectiveFlip:
   Now it toggles the perspective without altering the physical order.
*/
void intStackPerspectiveFlip(IntStackPerspective *sp) {
    if (strcmp(sp->perspective, "lifo") == 0) {
        intStackPerspectiveSetPerspective(sp, "fifo");
        printf("@%s perspective flipped to: FIFO\n", sp->name);
    } else {
        intStackPerspectiveSetPerspective(sp, "lifo");
        printf("@%s perspective flipped to: LIFO\n", sp->name);
    }
}

void freeIntStackPerspective(IntStackPerspective *sp) {
    if (sp) {
        free(sp);
    }
}
