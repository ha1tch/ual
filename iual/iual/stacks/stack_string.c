#include "stacks.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

StringStack *newStringStack(void) {
    StringStack *s = malloc(sizeof(StringStack));
    if (!s) { 
        perror("malloc newStringStack"); 
        exit(EXIT_FAILURE); 
    }
    s->capacity = 16;
    s->size = 0;
    s->data = malloc(sizeof(char*) * s->capacity);
    if (!s->data) { 
        perror("malloc newStringStack->data"); 
        exit(EXIT_FAILURE); 
    }
    strcpy(s->mode, "lifo");
    return s;
}

void stringStackPush(StringStack *s, const char *val) {
    if (s->size >= s->capacity) {
        s->capacity *= 2;
        char **newData = realloc(s->data, sizeof(char*) * s->capacity);
        if (!newData) { 
            perror("realloc in stringStackPush"); 
            exit(EXIT_FAILURE); 
        }
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

void freeStringStack(StringStack *s) {
    if (s) {
        for (int i = 0; i < s->size; i++) {
            free(s->data[i]);
        }
        free(s->data);
        free(s);
    }
}
