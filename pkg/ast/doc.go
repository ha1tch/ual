// Package ast defines the Abstract Syntax Tree nodes for ual programs.
//
// The AST represents the parsed structure of a ual program and is used by both
// the compiler (ual) and interpreter (iual). All node types implement the Node
// interface, with statements implementing Stmt and expressions implementing Expr.
//
// Key node types include:
//   - Program: root node containing all top-level statements
//   - StackDecl, ViewDecl: stack and view declarations
//   - FuncDecl, FuncCall: function definitions and calls
//   - StackOp, StackBlock: stack operations
//   - ComputeStmt, ConsiderStmt, SelectStmt: control constructs
//   - VarDecl, Assignment: variable handling
//   - IfStmt, WhileStmt, ForStmt: control flow
package ast
