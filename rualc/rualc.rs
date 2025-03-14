// ual_unified.rs
//
// A unified Chumsky-based parser for ual that includes:
//
// 1. Package/import declarations with export info.
// 2. Declarations (functions and global variables) and control flow constructs.
// 3. Expressions with operator precedence (extended unary, relational operators).
// 4. Data Constructors: table, array, and hash literals.
// 5. Result handling syntactic sugar (.consider { if_ok ... if_err ... }).
// 6. Stack operations: direct method calls, stacked mode syntax, and explicit stack creation.
// 7. Enhanced error recovery (using improved recovery combinators).
// 8. A stub semantic analysis phase for symbol resolution, scope tracking, and AST normalization.
//
// This is a foundation for a complete ual compiler.

use chumsky::prelude::*;
use chumsky::error::Simple;

// ---------- Custom Whitespace and Comments ----------

fn ws() -> impl Parser<char, (), Error = Simple<char>> {
    // Define comment parsers:
    let lua_comment = just("--").then(take_until(just('\n'))).padded();
    let cpp_comment = just("//").then(take_until(just('\n'))).padded();
    let c_comment = just("/*").then(take_until(just("*/"))).then_ignore(just("*/")).padded();
    // Combine standard whitespace and comments.
    text::whitespace()
        .or(lua_comment)
        .or(cpp_comment)
        .or(c_comment)
        .repeated()
        .map(|_| ())
}

// A helper to pad a parser with our custom whitespace.
fn padded<T, P>(parser: P) -> impl Parser<char, T, Error = Simple<char>>
where
    P: Parser<char, T, Error = Simple<char>>,
{
    parser.padded_by(ws(), ws())
}

// ---------- AST Definitions ----------

// Package & Import
#[derive(Debug, Clone, PartialEq)]
pub struct Program {
    pub package: PackageDecl,
    pub imports: Vec<ImportDecl>,
    pub decls: Vec<Decl>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct PackageDecl {
    pub name: String,
    pub exported: bool,
}

#[derive(Debug, Clone, PartialEq)]
pub struct ImportDecl {
    pub path: String,
}

// Declarations
#[derive(Debug, Clone, PartialEq)]
pub enum Decl {
    Function(FunctionDecl),
    GlobalVar(GlobalVarDecl),
}

#[derive(Debug, Clone, PartialEq)]
pub struct FunctionDecl {
    pub name: String,
    pub params: Vec<String>, // Parameters as names; types can be added later.
    pub body: Vec<Stmt>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct GlobalVarDecl {
    pub name: String,
    pub expr: Expr,
}

// Statements
#[derive(Debug, Clone, PartialEq)]
pub enum Stmt {
    Return(Option<Expr>),
    Expr(Expr),
    Assign(Vec<String>, Vec<Expr>),
    IfTrue { cond: Expr, block: Vec<Stmt> },
    IfFalse { cond: Expr, block: Vec<Stmt> },
    WhileTrue { cond: Expr, block: Vec<Stmt> },
    ForNum { var: String, start: Expr, end: Expr, step: Option<Expr>, block: Vec<Stmt> },
    ForGen { var: String, expr: Expr, block: Vec<Stmt> },
    Switch { expr: Expr, cases: Vec<Case>, default: Option<Vec<Stmt>> },
}

#[derive(Debug, Clone, PartialEq)]
pub struct Case {
    pub values: Vec<Expr>,
    pub block: Vec<Stmt>,
}

// Expressions
#[derive(Debug, Clone, PartialEq)]
pub enum Expr {
    Ident(String),
    Number(f64),
    String(String),
    Unary(String, Box<Expr>),
    Binary(Box<Expr>, String, Box<Expr>),
    Paren(Box<Expr>),
    // Data constructors:
    Table(Vec<TableField>),
    Array(Vec<Expr>),
    Hash(Vec<(Expr, Expr)>),
    // Result handling:
    ResultHandling { result: Box<Expr>, clauses: Vec<ResultHandlerClause> },
    // Explicit stack creation:
    StackCreation { args: Vec<Expr> },
}

#[derive(Debug, Clone, PartialEq)]
pub struct TableField {
    pub key: Option<Expr>,
    pub value: Expr,
}

#[derive(Debug, Clone, PartialEq)]
pub enum ResultHandlerClause {
    IfOk(Expr),
    IfErr(Expr),
}

// Stack operations
#[derive(Debug, Clone, PartialEq)]
pub enum StackOp {
    MethodCall { name: String, args: Vec<Expr> },
}

#[derive(Debug, Clone, PartialEq)]
pub struct StackedMode {
    pub target: Option<String>, // e.g., "rstack" or default "dstack"
    pub ops: Vec<StackOp>,
}

// ---------- Semantic Analysis Stub ----------

fn semantic_analysis(prog: Program) -> Program {
    println!("Performing semantic analysis (stub)...");
    // In a full implementation, this function would:
    //  - Build symbol tables and track scopes.
    //  - Mark identifiers as exported based on naming rules.
    //  - Transform legacy syntax (e.g., Forth-like stack ops) into a normalized AST.
    //  - Enrich AST nodes with type information.
    prog
}

// ---------- Parsers ----------

// -- Package and Import Parsers --

fn package_decl() -> impl Parser<char, PackageDecl, Error = Simple<char>> {
    just("package")
        .padded_by(ws(), ws())
        .ignore_then(text::ident().padded_by(ws(), ws()))
        .map(|name: String| PackageDecl {
            exported: name.chars().next().map(|c| c.is_uppercase()).unwrap_or(false),
            name,
        })
}

fn import_decl() -> impl Parser<char, ImportDecl, Error = Simple<char>> {
    just("import")
        .padded_by(ws(), ws())
        .ignore_then(string_literal())
        .map(|path| ImportDecl { path })
}

fn string_literal() -> impl Parser<char, String, Error = Simple<char>> {
    let inner = none_of("\"").repeated().collect::<String>();
    just('"')
        .ignore_then(inner)
        .then_ignore(just('"'))
        .padded_by(ws(), ws())
}

// -- Top-Level Declaration Parsers --

fn function_decl() -> impl Parser<char, Decl, Error = Simple<char>> {
    just("function")
        .padded_by(ws(), ws())
        .ignore_then(text::ident().padded_by(ws(), ws()))
        .then(
            text::ident()
                .separated_by(just(',').padded_by(ws(), ws()))
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
                .or_not()
                .map(|opt| opt.unwrap_or_else(Vec::new))
        )
        .then(block())
        .then_ignore(just("end").padded_by(ws(), ws()))
        .map(|((name, params), body)| Decl::Function(FunctionDecl { name, params, body }))
}

fn global_var_decl() -> impl Parser<char, Decl, Error = Simple<char>> {
    text::ident().padded_by(ws(), ws())
        .then_ignore(just('=').padded_by(ws(), ws()))
        .then(expr().padded_by(ws(), ws()))
        .map(|(name, expr)| Decl::GlobalVar(GlobalVarDecl { name, expr }))
}

fn top_level_decl() -> impl Parser<char, Decl, Error = Simple<char>> {
    choice((function_decl(), global_var_decl()))
}

// -- Program Parser with Enhanced Error Recovery --

fn program() -> impl Parser<char, Program, Error = Simple<char>> {
    package_decl()
        .then(import_decl().repeated())
        .then(top_level_decl().repeated())
        .map(|((pkg, imports), decls)| Program {
            package: pkg,
            imports,
            decls,
        })
        // Enhanced error recovery: if an error occurs, skip until end.
        .recover_with(skip_then_retry_until([], end()))
}

// -- Expression Parsers --

// Numeric literal parser: supports decimal, binary, and hexadecimal.
fn number_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    // Binary: 0b1010 or 0B1010
    let binary = just("0b")
        .or(just("0B"))
        .ignore_then(filter(|c: &char| *c == '0' || *c == '1').repeated().collect::<String>())
        .try_map(|s: String, span| {
            u64::from_str_radix(&s, 2)
                .map(|v| v as f64)
                .map_err(|e| Simple::custom(span, format!("Invalid binary literal: {}", e)))
        })
        .map(Expr::Number);
    // Hexadecimal: 0x1f or 0X1F
    let hex = just("0x")
        .or(just("0X"))
        .ignore_then(filter(|c: &char| c.is_digit(16)).repeated().collect::<String>())
        .try_map(|s: String, span| {
            u64::from_str_radix(&s, 16)
                .map(|v| v as f64)
                .map_err(|e| Simple::custom(span, format!("Invalid hexadecimal literal: {}", e)))
        })
        .map(Expr::Number);
    // Decimal: one or more digits, optionally with a fractional part.
    let decimal = text::int(10)
        .then(just('.').then(text::int(10)).or_not())
        .collect::<String>()
        .try_map(|s, span| {
            s.parse::<f64>()
                .map_err(|e| Simple::custom(span, format!("Invalid decimal literal: {}", e)))
        })
        .map(Expr::Number);
    choice((binary, hex, decimal))
}

fn ident_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    text::ident().map(Expr::Ident)
}

fn string_lit_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    let inner = none_of("\"").repeated().collect::<String>();
    just('"').ignore_then(inner).then_ignore(just('"')).map(Expr::String)
}

fn paren_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    expr().delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
        .map(|e| Expr::Paren(Box::new(e)))
}

fn primary_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    choice((number_expr(), string_lit_expr(), ident_expr(), paren_expr(), stack_creation_expr()))
}

// Explicit stack creation syntax: "Stack.new(" [ <expr-list> ] ")"
fn stack_creation_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    just("Stack.new")
        .padded_by(ws(), ws())
        .ignore_then(
            expr().separated_by(just(',').padded_by(ws(), ws()))
                .or_not()
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
        )
        .map(|opt_args| Expr::StackCreation { args: opt_args.unwrap_or_else(Vec::new) })
}

fn unary_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    let op_parser = choice((
         just('-').to("-".to_string()),
         just('!').to("!".to_string()),
         just('~').to("~".to_string()),
         just('+').to("+".to_string()),
    )).repeated();
    op_parser.then(primary_expr()).map(|(ops, expr)| {
        ops.into_iter().rev().fold(expr, |acc, op| Expr::Unary(op, Box::new(acc)))
    })
}

fn mul_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    unary_expr().then(
        (choice((just('*').to("*".to_string()), just('/').to("/".to_string())))
            .then(unary_expr()))
        .repeated()
    ).foldl(|lhs, (op, rhs)| Expr::Binary(Box::new(lhs), op, Box::new(rhs)))
}

fn add_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    mul_expr().then(
        (choice((just('+').to("+".to_string()), just('-').to("-".to_string())))
            .then(mul_expr()))
        .repeated()
    ).foldl(|lhs, (op, rhs)| Expr::Binary(Box::new(lhs), op, Box::new(rhs)))
}

fn shift_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    add_expr().then(
        (choice((just("<<").to("<<".to_string()), just(">>").to(">>".to_string())))
            .then(add_expr()))
        .repeated()
    ).foldl(|lhs, (op, rhs)| Expr::Binary(Box::new(lhs), op, Box::new(rhs)))
}

fn rel_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    shift_expr().then(
        (choice((
            just("<=").to("<=".to_string()),
            just(">=").to(">=".to_string()),
            just('<').to("<".to_string()),
            just('>').to(">".to_string()),
        )).then(shift_expr()))
        .repeated()
    ).foldl(|lhs, (op, rhs)| Expr::Binary(Box::new(lhs), op, Box::new(rhs)))
}

fn eq_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    rel_expr().then(
        (choice((just("==").to("==".to_string()), just("!=").to("!=".to_string())))
            .then(rel_expr()))
        .repeated()
    ).foldl(|lhs, (op, rhs)| Expr::Binary(Box::new(lhs), op, Box::new(rhs)))
}

fn bit_and_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    eq_expr().then(
        (just('&').to("&".to_string()).then(eq_expr())).repeated()
    ).foldl(|lhs, (op, rhs)| Expr::Binary(Box::new(lhs), op, Box::new(rhs)))
}

fn bit_xor_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    bit_and_expr().then(
        (just('^').to("^".to_string()).then(bit_and_expr())).repeated()
    ).foldl(|lhs, (op, rhs)| Expr::Binary(Box::new(lhs), op, Box::new(rhs)))
}

fn bit_or_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    bit_xor_expr().then(
        (just('|').to("|".to_string()).then(bit_xor_expr())).repeated()
    ).foldl(|lhs, (op, rhs)| Expr::Binary(Box::new(lhs), op, Box::new(rhs)))
}

fn expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    bit_or_expr()
}

// -- Data Constructors --

fn table_field() -> impl Parser<char, TableField, Error = Simple<char>> {
    let keydef = choice((
        text::ident().map(|s: String| Expr::Ident(s)).then_ignore(just('=').padded_by(ws(), ws())),
        expr().delimited_by(just('[').padded_by(ws(), ws()), just(']').padded_by(ws(), ws()))
            .then_ignore(just('=').padded_by(ws(), ws())),
    )).or_not();
    keydef.then(expr().padded_by(ws(), ws()))
         .map(|(key, value)| TableField { key, value })
}

fn table_constructor() -> impl Parser<char, Expr, Error = Simple<char>> {
    table_field()
        .separated_by(just(',').padded_by(ws(), ws()))
        .or_not()
        .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
        .map(|opt_fields| Expr::Table(opt_fields.unwrap_or_else(Vec::new)))
}

fn array_constructor() -> impl Parser<char, Expr, Error = Simple<char>> {
    expr()
        .separated_by(just(',').padded_by(ws(), ws()))
        .delimited_by(just('[').padded_by(ws(), ws()), just(']').padded_by(ws(), ws()))
        .map(Expr::Array)
}

fn key_value_pair() -> impl Parser<char, (Expr, Expr), Error = Simple<char>> {
    expr().padded_by(ws(), ws())
        .then_ignore(just('~').padded_by(ws(), ws()))
        .then(expr().padded_by(ws(), ws()))
}

fn hash_literal() -> impl Parser<char, Expr, Error = Simple<char>> {
    key_value_pair()
        .separated_by(just(',').padded_by(ws(), ws()))
        .or_not()
        .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
        .map(|opt_pairs| Expr::Hash(opt_pairs.unwrap_or_else(Vec::new)))
}

// -- Result Handling --

fn result_handler_clause() -> impl Parser<char, ResultHandlerClause, Error = Simple<char>> {
    let if_ok = just("if_ok")
        .padded_by(ws(), ws())
        .ignore_then(expr().padded_by(ws(), ws()))
        .map(ResultHandlerClause::IfOk);
    let if_err = just("if_err")
        .padded_by(ws(), ws())
        .ignore_then(expr().padded_by(ws(), ws()))
        .map(ResultHandlerClause::IfErr);
    choice((if_ok, if_err))
}

fn result_handler_block() -> impl Parser<char, Vec<ResultHandlerClause>, Error = Simple<char>> {
    result_handler_clause()
        .repeated()
        .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
}

fn result_handling_expr() -> impl Parser<char, Expr, Error = Simple<char>> {
    expr().then(
        just('.')
            .padded_by(ws(), ws())
            .ignore_then(just("consider"))
            .padded_by(ws(), ws())
            .ignore_then(result_handler_block())
            .or_not()
    ).map(|(base_expr, maybe_clauses)| {
         if let Some(clauses) = maybe_clauses {
             Expr::ResultHandling { result: Box::new(base_expr), clauses }
         } else {
             base_expr
         }
    })
}

// -- Stack Operations and Stacked Mode --

fn stack_op() -> impl Parser<char, StackOp, Error = Simple<char>> {
    let name = text::ident().padded_by(ws(), ws());
    // Literal parameter: colon then expression.
    let literal_param = just(':')
        .padded_by(ws(), ws())
        .ignore_then(expr().padded_by(ws(), ws()))
        .map(|arg| vec![arg]);
    // Parenthesized parameters.
    let paren_params = expr()
        .separated_by(just(',').padded_by(ws(), ws()))
        .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()));
    let params = literal_param.or(paren_params).or_not().map(|opt| opt.unwrap_or_else(Vec::new));
    name.then(params).map(|(name, args)| StackOp::MethodCall { name, args })
}

fn direct_stack_call() -> impl Parser<char, StackOp, Error = Simple<char>> {
    stack_op()
}

fn stacked_mode() -> impl Parser<char, StackedMode, Error = Simple<char>> {
    let selector = just('@')
        .padded_by(ws(), ws())
        .ignore_then(text::ident().padded_by(ws(), ws()))
        .or_not();
    let arrow = just('>').padded_by(ws(), ws());
    let ops = stack_op().padded_by(ws(), ws()).repeated();
    selector.then(arrow).then(ops).map(|((target, _), ops)| StackedMode { target, ops })
}

// -- Control Flow Parsers --

fn simple_stmt() -> impl Parser<char, Stmt, Error = Simple<char>> {
    expr().map(Stmt::Expr)
}

fn block() -> impl Parser<char, Vec<Stmt>, Error = Simple<char>> {
    simple_stmt()
        .repeated()
        .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
}

fn if_true_stmt() -> impl Parser<char, Stmt, Error = Simple<char>> {
    just("if_true")
        .padded_by(ws(), ws())
        .ignore_then(just('('))
        .ignore_then(expr().padded_by(ws(), ws()))
        .then_ignore(just(')').padded_by(ws(), ws()))
        .then(block())
        .then(just("end_if_true").or_not())
        .map(|(cond, block)| Stmt::IfTrue { cond, block })
}

fn if_false_stmt() -> impl Parser<char, Stmt, Error = Simple<char>> {
    just("if_false")
        .padded_by(ws(), ws())
        .ignore_then(just('('))
        .ignore_then(expr().padded_by(ws(), ws()))
        .then_ignore(just(')').padded_by(ws(), ws()))
        .then(block())
        .then(just("end_if_false").or_not())
        .map(|(cond, block)| Stmt::IfFalse { cond, block })
}

fn while_true_stmt() -> impl Parser<char, Stmt, Error = Simple<char>> {
    just("while_true")
        .padded_by(ws(), ws())
        .ignore_then(just('('))
        .ignore_then(expr().padded_by(ws(), ws()))
        .then_ignore(just(')').padded_by(ws(), ws()))
        .then(block())
        .then(just("end_while_true").or_not())
        .map(|(cond, block)| Stmt::WhileTrue { cond, block })
}

fn for_num_stmt() -> impl Parser<char, Stmt, Error = Simple<char>> {
    just("for")
        .padded_by(ws(), ws())
        .ignore_then(text::ident().padded_by(ws(), ws()))
        .then_ignore(just('=').padded_by(ws(), ws()))
        .then(expr().padded_by(ws(), ws()))
        .then_ignore(just(',').padded_by(ws(), ws()))
        .then(expr().padded_by(ws(), ws()))
        .then(just(',').padded_by(ws(), ws()).ignore_then(expr()).or_not())
        .then_ignore(just("do").padded_by(ws(), ws()))
        .then(block())
        .then_ignore(just("end").padded_by(ws(), ws()))
        .map(|(((var, start), end), step, block)| {
            Stmt::ForNum { var, start, end, step, block }
        })
}

fn for_gen_stmt() -> impl Parser<char, Stmt, Error = Simple<char>> {
    just("for")
        .padded_by(ws(), ws())
        .ignore_then(text::ident().padded_by(ws(), ws()))
        .then_ignore(just("in").padded_by(ws(), ws()))
        .then(expr().padded_by(ws(), ws()))
        .then_ignore(just("do").padded_by(ws(), ws()))
        .then(block())
        .then_ignore(just("end").padded_by(ws(), ws()))
        .map(|((var, expr_val), block)| Stmt::ForGen { var, expr: expr_val, block })
}

fn case_stmt() -> impl Parser<char, Case, Error = Simple<char>> {
    just("case")
        .padded_by(ws(), ws())
        .ignore_then(expr().separated_by(just(',').padded_by(ws(), ws())))
        .then_ignore(just(':').padded_by(ws(), ws()))
        .then(block())
        .map(|(values, block)| Case { values, block })
}

fn case_list() -> impl Parser<char, Vec<Case>, Error = Simple<char>> {
    case_stmt().repeated()
}

fn switch_stmt() -> impl Parser<char, Stmt, Error = Simple<char>> {
    just("switch_case")
        .padded_by(ws(), ws())
        .ignore_then(just('(').padded_by(ws(), ws()))
        .ignore_then(expr().padded_by(ws(), ws()))
        .then_ignore(just(')').padded_by(ws(), ws()))
        .then(
            case_list().then(
                just("default:")
                    .padded_by(ws(), ws())
                    .ignore_then(block())
                    .or_not()
            )
        )
        .then_ignore(just("end_switch").padded_by(ws(), ws()))
        .map(|(expr_val, (cases, default))| Stmt::Switch { expr: expr_val, cases, default })
}

// -- Top-Level Declaration and Program Parsers --

fn function_decl() -> impl Parser<char, Decl, Error = Simple<char>> {
    just("function")
        .padded_by(ws(), ws())
        .ignore_then(text::ident().padded_by(ws(), ws()))
        .then(
            text::ident()
                .separated_by(just(',').padded_by(ws(), ws()))
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
                .or_not()
                .map(|opt| opt.unwrap_or_else(Vec::new))
        )
        .then(block())
        .then_ignore(just("end").padded_by(ws(), ws()))
        .map(|((name, params), body)| Decl::Function(FunctionDecl { name, params, body }))
}

fn global_var_decl() -> impl Parser<char, Decl, Error = Simple<char>> {
    text::ident().padded_by(ws(), ws())
        .then_ignore(just('=').padded_by(ws(), ws()))
        .then(expr().padded_by(ws(), ws()))
        .map(|(name, expr)| Decl::GlobalVar(GlobalVarDecl { name, expr }))
}

fn top_level_decl() -> impl Parser<char, Decl, Error = Simple<char>> {
    choice((function_decl(), global_var_decl()))
}

fn program() -> impl Parser<char, Program, Error = Simple<char>> {
    package_decl()
        .then(import_decl().repeated())
        .then(top_level_decl().repeated())
        .map(|((pkg, imports), decls)| Program {
            package: pkg,
            imports,
            decls,
        })
        .recover_with(skip_then_retry_until([], end()))
}

// -- Enhanced Error Recovery: We add recovery strategies to key parsers as needed.
// (Here, we already use recover_with() on the program parser.)

// -- Unified Top-Level Parser --
fn unified_parser() -> impl Parser<char, Program, Error = Simple<char>> {
    program()
}

// ---------- Main (Testing Unified Parser) ----------

fn main() {
    let source = r#"
        package Main
        import "fmt"
        import "con"

        function Fibonacci(n) {
            if_true(n == 0) { return 1 } end_if_true
            return n + Fibonacci(n - 1)
        } end

        result = Fibonacci(5).consider { if_ok fmt.Printf("Success: %d", _1) if_err fmt.Printf("Error: %s", _1) };

        push(10);
        @rstack > push:42 swap;

        if_false(x) { y } end_if_false;
        while_true(z) { w } end_while_true;
        for i = start, end, step do { a } end;
        for item in iterator do { b } end;
        switch_case(val)
            case 1,2 : { c }
            case 3 : { d }
            default: { e }
        end_switch;
    "#;

    match unified_parser().then_ignore(end()).parse(source) {
        Ok(prog) => {
            println!("Parsed AST: {:#?}", prog);
            let normalized = semantic_analysis(prog);
            println!("Normalized AST: {:#?}", normalized);
        }
        Err(errors) => {
            println!("Errors during parsing:");
            for err in errors {
                println!("Error: {}", err);
            }
        }
    }
}
