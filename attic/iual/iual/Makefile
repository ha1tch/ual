# Makefile for iual project

CC = gcc
CFLAGS = -Wall -pthread -I. -Icontext -Iinterpreter -Ispawn -Istacks

TARGET = iual

# Directories
CONTEXT_DIR = context
INTERPRETER_DIR = interpreter
SPAWN_DIR = spawn
STACKS_DIR = stacks

# Object files (note: iual.c is the main file in the project root)
OBJS = iual.o \
       $(CONTEXT_DIR)/context.o \
       $(INTERPRETER_DIR)/interpreter.o \
       $(SPAWN_DIR)/spawn.o \
       $(STACKS_DIR)/stacks.o \
       $(STACKS_DIR)/stack_int.o \
       $(STACKS_DIR)/stack_string.o \
       $(STACKS_DIR)/stack_float.o

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $(TARGET) $(OBJS)

%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@

$(CONTEXT_DIR)/%.o: $(CONTEXT_DIR)/%.c
	$(CC) $(CFLAGS) -c $< -o $@

$(INTERPRETER_DIR)/%.o: $(INTERPRETER_DIR)/%.c
	$(CC) $(CFLAGS) -c $< -o $@

$(SPAWN_DIR)/%.o: $(SPAWN_DIR)/%.c
	$(CC) $(CFLAGS) -c $< -o $@

$(STACKS_DIR)/%.o: $(STACKS_DIR)/%.c
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -f $(TARGET) $(OBJS)
