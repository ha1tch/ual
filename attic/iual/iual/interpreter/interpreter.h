#ifndef INTERPRETER_H
#define INTERPRETER_H

// Processes a compound command line (e.g., "@dstack: push:1 pop add")
void processCompoundCommand(char *input);

// Executes a spawn command (used by spawn script executor)
void executeSpawnCommand(const char *cmd);

#endif
