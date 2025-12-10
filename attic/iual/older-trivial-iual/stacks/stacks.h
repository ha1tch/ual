#ifndef STACKS_H
#define STACKS_H

// ------------------------------
// IntStack Definition and Prototypes
// ------------------------------

typedef struct {
    int *data;
    int size;
    int capacity;
    char mode[5]; // "lifo" or "fifo"
} IntStack;

// Basic operations for IntStack
IntStack *newIntStack();
void intStackPush(IntStack *s, int val);
int intStackPop(IntStack *s, int *success);
int intStackDepth(IntStack *s);
void intStackPrint(IntStack *s);
void intStackSetMode(IntStack *s, const char *mode);
void intStackFlip(IntStack *s);

// Forth-like operations for IntStack
int intStackAdd(IntStack *s);         // ( a b -- a+b )
int intStackSub(IntStack *s);         // ( a b -- a-b )
int intStackMul(IntStack *s);         // ( a b -- a*b )
int intStackDiv(IntStack *s);         // ( a b -- a/b )
int intStackTuck(IntStack *s);        // ( a b -- b a b )
int intStackPick(IntStack *s, int n); // ( ... x_n ... x_0 n -- ... x_n ... x_0 x_n )
int intStackRoll(IntStack *s, int n); // ( ... x_n ... x_0 n -- ... x_1 x_0 x_n )
int intStackOver2(IntStack *s);       // ( a b c d -- a b c d a b )
int intStackDrop2(IntStack *s);       // ( a b c d -- a b )
int intStackSwap2(IntStack *s);       // ( a b c d -- c d a b )
int intStackDepthValue(IntStack *s);  // Returns current stack depth

// Memory operations for IntStack
int intStackStore(IntStack *s);       // ( value address -- )
int intStackLoad(IntStack *s);        // ( address -- value )

// Bitwise operations for IntStack
int intStackAnd(IntStack *s);
int intStackOr(IntStack *s);
int intStackXor(IntStack *s);
int intStackShl(IntStack *s);
int intStackShr(IntStack *s);

// ------------------------------
// StringStack Definition and Prototypes
// ------------------------------

typedef struct {
    char **data;
    int size;
    int capacity;
    char mode[5]; // "lifo" or "fifo"
} StringStack;

// Basic operations for StringStack
StringStack *newStringStack();
void stringStackPush(StringStack *s, const char *val);
char *stringStackPop(StringStack *s, int *success);
int stringStackDepth(StringStack *s);
void stringStackPrint(StringStack *s);
void stringStackSetMode(StringStack *s, const char *mode);
void stringStackFlip(StringStack *s);

// Forth-like operations for StringStack
int stringStackAdd(StringStack *s);                     // Concatenates the top two strings
int stringStackSub(StringStack *s, const char *trimChar); // Trims trailing occurrences of trimChar
int stringStackMul(StringStack *s, int n);              // Replicates the top string n times
int stringStackDiv(StringStack *s, const char *delim);    // Splits the top string on delim and joins with a space

// ------------------------------
// Cleanup Functions for Stacks
// ------------------------------

void freeIntStack(IntStack *s);
void freeStringStack(StringStack *s);

// ------------------------------
// New: IntStackPerspective Definition and Prototypes
// ------------------------------

typedef struct {
    IntStack *physical;         // Underlying physical stack
    int startIndex;             // Logical front pointer for FIFO mode
    char perspective[10];       // Current perspective ("lifo", "fifo", etc.)
} IntStackPerspective;

IntStackPerspective *newIntStackPerspective(IntStack *physical);
void freeIntStackPerspective(IntStackPerspective *sp);
void intStackPerspectivePush(IntStackPerspective *sp, int val);
int intStackPerspectivePop(IntStackPerspective *sp, int *success);
void intStackPerspectiveSetPerspective(IntStackPerspective *sp, const char *persp);
void intStackPerspectiveFlip(IntStackPerspective *sp);
void intStackPerspectivePrint(IntStackPerspective *sp);

#endif
