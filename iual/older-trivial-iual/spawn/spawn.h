#ifndef SPAWN_H
#define SPAWN_H

#include <pthread.h>

typedef struct Spawn {
    char *name;
    pthread_t thread;
    pthread_mutex_t mtx;
    pthread_cond_t cond;
    int paused;
    int stop;
    char *script; // holds a multi-line script
} Spawn;

Spawn *newSpawn(const char *name);
void spawnPause(Spawn *sp);
void spawnResume(Spawn *sp);
void spawnStop(Spawn *sp);
void spawnRunScript(Spawn *sp);

#endif
