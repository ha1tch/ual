#include "stacks.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

FloatStack *newFloatStack(void) {
    FloatStack *s = malloc(sizeof(FloatStack));
    if (!s) { 
        perror("malloc newFloatStack"); 
        exit(EXIT_FAILURE); 
    }
    s->capacity = 16;
    s->size = 0;
    s->data = malloc(sizeof(double) * s->capacity);
    if (!s->data) { 
        perror("malloc newFloatStack->data"); 
        exit(EXIT_FAILURE); 
    }
    strcpy(s->mode, "lifo");
    return s;
}

void floatStackPush(FloatStack *s, double val) {
    if (s->size >= s->capacity) {
        s->capacity *= 2;
        double *newData = realloc(s->data, sizeof(double) * s->capacity);
        if (!newData) { 
            perror("realloc in floatStackPush"); 
            exit(EXIT_FAILURE); 
        }
        s->data = newData;
    }
    s->data[s->size++] = val;
}

double floatStackPop(FloatStack *s, int *success) {
    if (s->size == 0) {
        *success = 0;
        return 0.0;
    }
    if (strcmp(s->mode, "fifo") == 0) {
        double val = s->data[0];
        memmove(s->data, s->data + 1, sizeof(double) * (s->size - 1));
        s->size--;
        *success = 1;
        return val;
    }
    *success = 1;
    return s->data[--s->size];
}

int floatStackDepth(FloatStack *s) {
    return s->size;
}

void floatStackPrint(FloatStack *s) {
    printf("FloatStack (%s mode): ", s->mode);
    for (int i = 0; i < s->size; i++) {
        printf("%f ", s->data[i]);
    }
    printf("\n");
}

void floatStackSetMode(FloatStack *s, const char *mode) {
    if (strcmp(mode, "lifo") == 0 || strcmp(mode, "fifo") == 0)
        strcpy(s->mode, mode);
}

void floatStackFlip(FloatStack *s) {
    for (int i = 0, j = s->size - 1; i < j; i++, j--) {
        double tmp = s->data[i];
        s->data[i] = s->data[j];
        s->data[j] = tmp;
    }
    if (strcmp(s->mode, "lifo") == 0)
        strcpy(s->mode, "fifo");
    else
        strcpy(s->mode, "lifo");
}

void freeFloatStack(FloatStack *s) {
    if (s) {
        free(s->data);
        free(s);
    }
}
