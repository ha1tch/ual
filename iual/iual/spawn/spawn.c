#include "spawn/spawn.h"
#include "interpreter/interpreter.h"  // For executeSpawnCommand()
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>

static void *spawnMain(void *arg);

Spawn *newSpawn(const char *name) {
    Spawn *sp = malloc(sizeof(Spawn));
    if (!sp) { perror("malloc newSpawn"); exit(EXIT_FAILURE); }
    sp->name = strdup(name);
    sp->paused = 0;
    sp->stop = 0;
    sp->script = NULL;
    pthread_mutex_init(&sp->mtx, NULL);
    pthread_cond_init(&sp->cond, NULL);
    if (pthread_create(&sp->thread, NULL, spawnMain, sp) != 0) {
        perror("pthread_create in newSpawn");
        exit(EXIT_FAILURE);
    }
    return sp;
}

static void *spawnMain(void *arg) {
    Spawn *sp = (Spawn *)arg;
    while (1) {
        pthread_mutex_lock(&sp->mtx);
        // Wait until a script is available or stop is signaled.
        while (!sp->stop && sp->script == NULL)
            pthread_cond_wait(&sp->cond, &sp->mtx);
        if (sp->stop) {
            pthread_mutex_unlock(&sp->mtx);
            break;
        }
        char *script = sp->script;
        sp->script = NULL;
        pthread_mutex_unlock(&sp->mtx);

        // Execute the script: split into lines and dispatch each line.
        printf("[spawn:%s] Executing script:\n%s\n", sp->name, script);
        char *line = strtok(script, "\n");
        while (line != NULL) {
            executeSpawnCommand(line);
            line = strtok(NULL, "\n");
        }
        free(script);
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
