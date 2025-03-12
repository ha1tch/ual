#ifndef STACKS_H
#define STACKS_H

/* ---------- Common Global Memory ---------- */
extern int globalMemory[1024];

/* ---------- Global IntStack List for Perspectives ---------- */
typedef struct IntStackNode {
    char *name;
    struct IntStackPerspective *persp;
    struct IntStackNode *next;
} IntStackNode;
extern IntStackNode *globalIntStackList;

/* Prototype for lookup function */
struct IntStackPerspective *findIntStackPerspective(const char *name);

/* ---------- IntStack Definitions ---------- */
typedef struct {
    int *data;
    int size;
    int capacity;
    char mode[10]; // "lifo" or "fifo"
} IntStack;

IntStack *newIntStack(void);
void intStackPush(IntStack *s, int val);
int intStackPop(IntStack *s, int *success);
int intStackDepth(IntStack *s);
void intStackPrint(IntStack *s);
void intStackSetMode(IntStack *s, const char *mode);
void intStackFlip(IntStack *s);
void freeIntStack(IntStack *s);

/* Forth‑like operations for IntStack */
int intStackAdd(IntStack *s);
int intStackSub(IntStack *s);
int intStackMul(IntStack *s);
int intStackDiv(IntStack *s);
int intStackTuck(IntStack *s);
int intStackPick(IntStack *s, int n);
int intStackRoll(IntStack *s, int n);
int intStackOver2(IntStack *s);
int intStackDrop2(IntStack *s);
int intStackSwap2(IntStack *s);
int intStackDepthValue(IntStack *s);

/* Memory operations for IntStack */
int intStackStore(IntStack *s);
int intStackLoad(IntStack *s);

/* Bitwise operations for IntStack */
int intStackAnd(IntStack *s);
int intStackOr(IntStack *s);
int intStackXor(IntStack *s);
int intStackShl(IntStack *s);
int intStackShr(IntStack *s);

/* ---------- IntStackPerspective Definitions ---------- */
struct IntStackPerspective {
    IntStack *physical;         // Underlying physical int stack.
    int startIndex;             // Logical front pointer (for FIFO).
    char perspective[10];       // "lifo" or "fifo"
    char name[32];              // The stack's name (e.g. "dstack", "rstack")
};

typedef struct IntStackPerspective IntStackPerspective;

IntStackPerspective *newIntStackPerspective(IntStack *s, const char *name);
void intStackPerspectivePush(IntStackPerspective *sp, int val);
int intStackPerspectivePop(IntStackPerspective *sp, int *success);
void intStackPerspectivePrint(IntStackPerspective *sp);
void intStackPerspectiveSetPerspective(IntStackPerspective *sp, const char *mode);
void intStackPerspectiveFlip(IntStackPerspective *sp);
void freeIntStackPerspective(IntStackPerspective *sp);

/* ---------- StringStack Definitions ---------- */
typedef struct {
    char **data;
    int size;
    int capacity;
    char mode[10]; // "lifo" or "fifo"
} StringStack;

StringStack *newStringStack(void);
void stringStackPush(StringStack *s, const char *val);
char *stringStackPop(StringStack *s, int *success);
int stringStackDepth(StringStack *s);
void stringStackPrint(StringStack *s);
void stringStackSetMode(StringStack *s, const char *mode);
void stringStackFlip(StringStack *s);
int stringStackAdd(StringStack *s);
int stringStackSub(StringStack *s, const char *trimChar);
int stringStackMul(StringStack *s, int n);
int stringStackDiv(StringStack *s, const char *delim);
void freeStringStack(StringStack *s);

/* ---------- FloatStack Definitions ---------- */
typedef struct {
    double *data;
    int size;
    int capacity;
    char mode[10]; // "lifo" or "fifo"
} FloatStack;

FloatStack *newFloatStack(void);
void floatStackPush(FloatStack *s, double val);
double floatStackPop(FloatStack *s, int *success);
int floatStackDepth(FloatStack *s);
void floatStackPrint(FloatStack *s);
void floatStackSetMode(FloatStack *s, const char *mode);
void floatStackFlip(FloatStack *s);
void freeFloatStack(FloatStack *s);

#endif
