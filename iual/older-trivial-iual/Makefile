CC = gcc
CFLAGS = -Wall -pthread -I.
OBJS = iual.o context.o stacks/stacks.o spawn/spawn.o interpreter/interpreter.o

iual: $(OBJS)
	$(CC) $(CFLAGS) -o iual $(OBJS)

iual.o: iual.c stacks/stacks.h spawn/spawn.h interpreter/interpreter.h
	$(CC) $(CFLAGS) -c iual.c

stacks/stacks.o: stacks/stacks.c stacks/stacks.h
	$(CC) $(CFLAGS) -c stacks/stacks.c -o stacks/stacks.o

spawn/spawn.o: spawn/spawn.c spawn/spawn.h
	$(CC) $(CFLAGS) -c spawn/spawn.c -o spawn/spawn.o

interpreter/interpreter.o: interpreter/interpreter.c interpreter/interpreter.h stacks/stacks.h spawn/spawn.h
	$(CC) $(CFLAGS) -c interpreter/interpreter.c -o interpreter/interpreter.o

clean:
	rm -f iual $(OBJS)
