#include "interpreter/interpreter.h"
#include "stacks/stacks.h"
#include "spawn/spawn.h"
#include <ctype.h>      // For isspace, isalpha, isdigit.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* --- Helper Functions --- */

// Trim leading and trailing whitespace.
static char *trimWhitespace(char *str) {
    while (isspace(*str)) str++;
    char *end = str + strlen(str) - 1;
    while (end > str && isspace(*end)) { *end = '\0'; end--; }
    return str;
}

/* --- Value Type Definitions --- */
typedef enum {
    VAL_INT,
    VAL_STRING
} ValueType;

typedef struct {
    ValueType type;
    int intValue;
    char stringValue[128];
} Value;

static Value makeIntValue(int v) {
    Value val;
    val.type = VAL_INT;
    val.intValue = v;
    return val;
}

static Value makeStringValue(const char *s) {
    Value val;
    val.type = VAL_STRING;
    strncpy(val.stringValue, s, sizeof(val.stringValue) - 1);
    val.stringValue[sizeof(val.stringValue) - 1] = '\0';
    return val;
}

/* --- AST Definitions --- */
typedef enum {
    AST_LITERAL,
    AST_IDENTIFIER,
    AST_FUNCTION_CALL
} ASTNodeType;

typedef enum {
    LITERAL_INT,
    LITERAL_STRING
} LiteralType;

typedef struct ASTNode {
    ASTNodeType type;
    LiteralType literalType; // Valid if type == AST_LITERAL
    char lexeme[128];        // For identifiers and function names.
    int intLiteral;          // If numeric literal.
    char stringLiteral[128]; // If string literal.
    struct ASTNode **arguments; // For function calls.
    int argCount;
} ASTNode;

static ASTNode *newASTNode(ASTNodeType type) {
    ASTNode *node = malloc(sizeof(ASTNode));
    if (!node) { perror("malloc ASTNode"); exit(EXIT_FAILURE); }
    node->type = type;
    node->arguments = NULL;
    node->argCount = 0;
    return node;
}

static void freeAST(ASTNode *node) {
    if (!node) return;
    for (int i = 0; i < node->argCount; i++) {
        freeAST(node->arguments[i]);
    }
    free(node->arguments);
    free(node);
}

/* --- Lexer Definitions --- */
// Note: The colon ':' is now treated as part of identifiers.
typedef enum {
    TOKEN_IDENTIFIER,
    TOKEN_NUMBER,
    TOKEN_STRING,
    TOKEN_SYMBOL,
    TOKEN_END
} TokenType;

typedef struct {
    TokenType type;
    char lexeme[128];
} Token;

typedef struct {
    const char *start;
    const char *current;
} Lexer;

static void initLexer(Lexer *lexer, const char *source) {
    lexer->start = source;
    lexer->current = source;
}

static int isAtEnd(Lexer *lexer) {
    return *lexer->current == '\0';
}

static char advance(Lexer *lexer) {
    return *(lexer->current++);
}

static char peek(Lexer *lexer) {
    return *lexer->current;
}

static void skipWhitespace(Lexer *lexer) {
    while (!isAtEnd(lexer) && isspace(peek(lexer)))
        advance(lexer);
}

static void stringCopy(char *dest, const char *start, size_t len) {
    strncpy(dest, start, len);
    dest[len] = '\0';
}

static Token makeToken(Lexer *lexer, TokenType type) {
    Token token;
    token.type = type;
    size_t len = lexer->current - lexer->start;
    if (len > sizeof(token.lexeme) - 1) len = sizeof(token.lexeme) - 1;
    stringCopy(token.lexeme, lexer->start, len);
    return token;
}

static int isAlpha(char c) {
    return isalpha(c) || c == '_' || c == ':';
}

static Token lexToken(Lexer *lexer) {
    skipWhitespace(lexer);
    lexer->start = lexer->current;
    if (isAtEnd(lexer))
        return makeToken(lexer, TOKEN_END);
    char c = advance(lexer);
    if (isAlpha(c)) {
        while (!isAtEnd(lexer) && (isAlpha(peek(lexer)) || isdigit(peek(lexer))))
            advance(lexer);
        return makeToken(lexer, TOKEN_IDENTIFIER);
    }
    if (isdigit(c)) {
        while (!isAtEnd(lexer) && isdigit(peek(lexer)))
            advance(lexer);
        return makeToken(lexer, TOKEN_NUMBER);
    }
    if (c == '"' || c == '\'') {
        char quote = c;
        while (!isAtEnd(lexer) && peek(lexer) != quote)
            advance(lexer);
        if (!isAtEnd(lexer)) advance(lexer);
        return makeToken(lexer, TOKEN_STRING);
    }
    if (strchr("(),@", c))
        return makeToken(lexer, TOKEN_SYMBOL);
    return makeToken(lexer, TOKEN_IDENTIFIER);
}

/* --- Parser Definitions --- */
typedef struct {
    Lexer lexer;
    Token current;
} Parser;

static void initParser(Parser *parser, const char *source) {
    initLexer(&parser->lexer, source);
    parser->current = lexToken(&parser->lexer);
}

static void advanceParser(Parser *parser) {
    parser->current = lexToken(&parser->lexer);
}

static int match(Parser *parser, const char *expected) {
    if (parser->current.type == TOKEN_SYMBOL && strcmp(parser->current.lexeme, expected) == 0) {
        advanceParser(parser);
        return 1;
    }
    return 0;
}

static ASTNode *parseExpression(Parser *parser);

static ASTNode *parsePrimary(Parser *parser) {
    ASTNode *node = NULL;
    if (parser->current.type == TOKEN_NUMBER) {
        node = newASTNode(AST_LITERAL);
        node->literalType = LITERAL_INT;
        node->intLiteral = atoi(parser->current.lexeme);
        advanceParser(parser);
    } else if (parser->current.type == TOKEN_STRING) {
        node = newASTNode(AST_LITERAL);
        node->literalType = LITERAL_STRING;
        char *str = parser->current.lexeme;
        size_t len = strlen(str);
        if (len >= 2) {
            strncpy(node->stringLiteral, str + 1, len - 2);
            node->stringLiteral[len - 2] = '\0';
        } else {
            node->stringLiteral[0] = '\0';
        }
        advanceParser(parser);
    } else if (parser->current.type == TOKEN_IDENTIFIER) {
        char *colonPos = strchr(parser->current.lexeme, ':');
        if (colonPos) {
            int opLen = colonPos - parser->current.lexeme;
            char opName[32];
            if (opLen > 31) opLen = 31;
            strncpy(opName, parser->current.lexeme, opLen);
            opName[opLen] = '\0';
            char argPart[128];
            strcpy(argPart, colonPos + 1);
            ASTNode *callNode = newASTNode(AST_FUNCTION_CALL);
            strncpy(callNode->lexeme, opName, sizeof(callNode->lexeme) - 1);
            callNode->argCount = 1;
            callNode->arguments = malloc(sizeof(ASTNode*));
            ASTNode *argNode = newASTNode(AST_LITERAL);
            if (isdigit(argPart[0]) || (argPart[0]=='-' && isdigit(argPart[1]))) {
                argNode->literalType = LITERAL_INT;
                argNode->intLiteral = atoi(argPart);
            } else {
                argNode->literalType = LITERAL_STRING;
                strncpy(argNode->stringLiteral, argPart, sizeof(argNode->stringLiteral) - 1);
                argNode->stringLiteral[sizeof(argNode->stringLiteral) - 1] = '\0';
            }
            callNode->arguments[0] = argNode;
            advanceParser(parser);
            node = callNode;
        } else {
            node = newASTNode(AST_IDENTIFIER);
            strncpy(node->lexeme, parser->current.lexeme, sizeof(node->lexeme)-1);
            advanceParser(parser);
        }
    } else if (parser->current.type == TOKEN_SYMBOL && parser->current.lexeme[0]=='@') {
        node = newASTNode(AST_IDENTIFIER);
        strncpy(node->lexeme, parser->current.lexeme, sizeof(node->lexeme)-1);
        advanceParser(parser);
    } else if (match(parser, "(")) {
        node = parseExpression(parser);
        if (!match(parser, ")"))
            printf("Error: expected ')'\n");
    } else {
        printf("Unexpected token: %s\n", parser->current.lexeme);
        advanceParser(parser);
    }
    // Auto-wrap known commands as function calls if not already a call.
    if (node && node->type == AST_IDENTIFIER) {
        if (strcmp(node->lexeme, "push") == 0 ||
            strcmp(node->lexeme, "pop") == 0 ||
            strcmp(node->lexeme, "add") == 0 ||
            strcmp(node->lexeme, "sub") == 0 ||
            strcmp(node->lexeme, "mul") == 0 ||
            strcmp(node->lexeme, "div") == 0 ||
            strcmp(node->lexeme, "flip") == 0 ||
            strcmp(node->lexeme, "print") == 0 ||
            strcmp(node->lexeme, "bring") == 0 ||
            strcmp(node->lexeme, "lifo") == 0 ||
            strcmp(node->lexeme, "fifo") == 0) {
            ASTNode *callNode = newASTNode(AST_FUNCTION_CALL);
            strncpy(callNode->lexeme, node->lexeme, sizeof(callNode->lexeme)-1);
            callNode->argCount = 0;
            callNode->arguments = NULL;
            freeAST(node);
            node = callNode;
        }
    }
    return node;
}

static ASTNode *parseFunctionCall(Parser *parser, ASTNode *callee) {
    ASTNode *node = newASTNode(AST_FUNCTION_CALL);
    strncpy(node->lexeme, callee->lexeme, sizeof(node->lexeme)-1);
    freeAST(callee);
    node->argCount = 0;
    node->arguments = NULL;
    if (!match(parser, ")")) {
        do {
            ASTNode *arg = parseExpression(parser);
            node->argCount++;
            node->arguments = realloc(node->arguments, sizeof(ASTNode*) * node->argCount);
            node->arguments[node->argCount - 1] = arg;
        } while (match(parser, ","));
        if (!match(parser, ")"))
            printf("Error: expected ')'\n");
    }
    return node;
}

static ASTNode *parseExpression(Parser *parser) {
    ASTNode *node = parsePrimary(parser);
    while (parser->current.type == TOKEN_SYMBOL && strcmp(parser->current.lexeme, "(") == 0) {
        advanceParser(parser);
        node = parseFunctionCall(parser, node);
    }
    return node;
}

/* --- Evaluator Definitions --- */
static Value evaluateAST(ASTNode *node, int *success);

static Value evalAdd(Value a, Value b) {
    Value result;
    if (a.type == VAL_INT && b.type == VAL_INT) {
        result = makeIntValue(a.intValue + b.intValue);
    } else {
        char bufA[128], bufB[128];
        if (a.type == VAL_INT)
            snprintf(bufA, sizeof(bufA), "%d", a.intValue);
        else
            strncpy(bufA, a.stringValue, sizeof(bufA)-1);
        if (b.type == VAL_INT)
            snprintf(bufB, sizeof(bufB), "%d", b.intValue);
        else
            strncpy(bufB, b.stringValue, sizeof(bufB)-1);
        char concat[256];
        snprintf(concat, sizeof(concat), "%s%s", bufA, bufB);
        result = makeStringValue(concat);
    }
    return result;
}

static Value evalSub(Value a, Value b) {
    Value result = makeIntValue(a.intValue - b.intValue);
    return result;
}

static Value evalMul(Value a, Value b) {
    Value result = makeIntValue(a.intValue * b.intValue);
    return result;
}

static Value evalDiv(Value a, Value b) {
    if (b.intValue == 0) {
        printf("Division by zero\n");
        return makeIntValue(0);
    }
    Value result = makeIntValue(a.intValue / b.intValue);
    return result;
}

static Value evaluateFunctionCall(ASTNode *node, int *success, const char *currentStackSelector, IntStackPerspective *currentStack) {
    Value result = makeIntValue(0);
    Value *args = NULL;
    for (int i = 0; i < node->argCount; i++) {
        Value val = evaluateAST(node->arguments[i], success);
        if (!(*success)) break;
        args = realloc(args, sizeof(Value) * (i + 1));
        args[i] = val;
    }
    if (strcmp(node->lexeme, "push") == 0) {
        for (int i = 0; i < node->argCount; i++) {
            if (args[i].type == VAL_INT)
                intStackPerspectivePush(currentStack, args[i].intValue);
            else
                printf("push: string value not supported on int stack\n");
        }
    } else if (strcmp(node->lexeme, "pop") == 0) {
        int ok;
        int val = intStackPerspectivePop(currentStack, &ok);
        if (ok)
            result = makeIntValue(val);
        else {
            printf("Stack is empty\n");
            *success = 0;
        }
    } else if (strcmp(node->lexeme, "add") == 0) {
        if (node->argCount == 0) {
            if (!intStackAdd(currentStack->physical))
                *success = 0;
        } else {
            result = evalAdd(args[0], args[1]);
        }
    } else if (strcmp(node->lexeme, "sub") == 0) {
        if (node->argCount == 0) {
            if (!intStackSub(currentStack->physical))
                *success = 0;
        } else {
            result = evalSub(args[0], args[1]);
        }
    } else if (strcmp(node->lexeme, "mul") == 0) {
        if (node->argCount == 0) {
            if (!intStackMul(currentStack->physical))
                *success = 0;
        } else {
            result = evalMul(args[0], args[1]);
        }
    } else if (strcmp(node->lexeme, "div") == 0) {
        if (node->argCount == 0) {
            if (!intStackDiv(currentStack->physical))
                *success = 0;
        } else {
            result = evalDiv(args[0], args[1]);
        }
    } else if (strcmp(node->lexeme, "flip") == 0) {
        intStackPerspectiveFlip(currentStack);
    } else if (strcmp(node->lexeme, "print") == 0) {
        intStackPerspectivePrint(currentStack);
    } else if (strcmp(node->lexeme, "bring") == 0) {
        if (node->argCount < 2) {
            printf("bring requires two arguments\n");
            *success = 0;
        } else {
            char srcStackName[32];
            strncpy(srcStackName, node->arguments[1]->lexeme, sizeof(srcStackName)-1);
            srcStackName[sizeof(srcStackName)-1] = '\0';
            if (srcStackName[0] == '@') {
                memmove(srcStackName, srcStackName + 1, strlen(srcStackName));
            }
            IntStackPerspective *srcPersp = findIntStackPerspective(srcStackName);
            if (!srcPersp) {
                printf("No int stack named '%s' found.\n", srcStackName);
                *success = 0;
            } else {
                int ok;
                int value = intStackPerspectivePop(srcPersp, &ok);
                if (!ok) {
                    printf("Source stack '%s' is empty.\n", srcStackName);
                    *success = 0;
                } else {
                    intStackPerspectivePush(currentStack, value);
                    printf("Brought value from int stack '%s' to selected stack '%s'\n", srcStackName, currentStackSelector);
                }
            }
        }
    } else if (strcmp(node->lexeme, "lifo") == 0) {
        intStackPerspectiveSetPerspective(currentStack, "lifo");
        printf("@%s perspective set: LIFO\n", currentStackSelector);
    } else if (strcmp(node->lexeme, "fifo") == 0) {
        intStackPerspectiveSetPerspective(currentStack, "fifo");
        printf("@%s perspective set: FIFO\n", currentStackSelector);
    } else {
        printf("Unknown function call: %s\n", node->lexeme);
        *success = 0;
    }
    free(args);
    return result;
}

static Value evaluateAST(ASTNode *node, int *success) {
    if (!node) { *success = 0; return makeIntValue(0); }
    if (node->type == AST_LITERAL) {
        if (node->literalType == LITERAL_INT)
            return makeIntValue(node->intLiteral);
        else
            return makeStringValue(node->stringLiteral);
    } else if (node->type == AST_IDENTIFIER) {
        return makeIntValue(atoi(node->lexeme));
    } else if (node->type == AST_FUNCTION_CALL) {
        extern char currentStackSelector[32];
        extern IntStackPerspective *currentStackPerspective;
        return evaluateFunctionCall(node, success, currentStackSelector, currentStackPerspective);
    }
    *success = 0;
    return makeIntValue(0);
}

/* --- Global Evaluation Context --- */
char currentStackSelector[32] = "";
IntStackPerspective *currentStackPerspective = NULL;

void processCompoundCommand(char *input) {
    // Expected format: "@<selector>: <command tokens>"
    char *colon = strchr(input, ':');
    if (!colon) {
        printf("Invalid compound command format.\n");
        return;
    }
    *colon = '\0';
    char *selector = input + 1;  // Skip '@'
    char *commands = colon + 1;
    commands = trimWhitespace(commands);
    
    IntStackPerspective *stackPerspective = findIntStackPerspective(selector);
    if (!stackPerspective) {
        printf("No int stack named '%s' found.\n", selector);
        return;
    }
    strncpy(currentStackSelector, selector, sizeof(currentStackSelector) - 1);
    currentStackSelector[sizeof(currentStackSelector) - 1] = '\0';
    currentStackPerspective = stackPerspective;
    
    Parser parser;
    initParser(&parser, commands);
    while (parser.current.type != TOKEN_END) {
        ASTNode *node = parseExpression(&parser);
        int succ = 1;
        Value val = evaluateAST(node, &succ);
        if (node->type == AST_LITERAL || node->type == AST_IDENTIFIER) {
            if (val.type == VAL_INT)
                intStackPerspectivePush(stackPerspective, val.intValue);
            else
                printf("Literal string value not supported on int stack\n");
        }
        freeAST(node);
    }
}

void executeSpawnCommand(const char *cmd) {
    char *copy = strdup(cmd);
    char *token = strtok(copy, " ");
    if (!token) { 
        free(copy);
        return;
    }
    if (strcmp(token, "list") == 0) {
        printf("[spawn] list command executed.\n");
    } else if (strcmp(token, "add") == 0) {
        token = strtok(NULL, " ");
        if (token)
            printf("[spawn] add command: %s\n", token);
    } else if (strcmp(token, "pause") == 0) {
        token = strtok(NULL, " ");
        if (token)
            printf("[spawn] pause command: %s\n", token);
    } else if (strcmp(token, "resume") == 0) {
        token = strtok(NULL, " ");
        if (token)
            printf("[spawn] resume command: %s\n", token);
    } else if (strcmp(token, "stop") == 0) {
        token = strtok(NULL, " ");
        if (token)
            printf("[spawn] stop command: %s\n", token);
    } else {
        printf("[spawn] Unknown command: %s\n", token);
    }
    free(copy);
}
