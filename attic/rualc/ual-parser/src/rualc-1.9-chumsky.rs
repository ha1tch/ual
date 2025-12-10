// complete_ual_parser.rs
//
// A comprehensive Chumsky-based parser for the ual programming language that supports:
// 1. All features from ual 1.3 through 1.9 proposals
// 2. Rich error handling and recovery
// 3. Detailed location information for diagnostics
// 4. Semantic analysis with symbol resolution
// 5. Support for modern ual features including:
//    - Stack perspectives (LIFO, FIFO, MAXFO, MINFO, Hashed)
//    - Ownership and borrowing (Owned, Borrowed, Mutable)
//    - Generalized pattern matching
//    - Crosstack operations with the tilde operator
//    - Hash literals with tilde separator
//    - Defer statements for resource management

use chumsky::prelude::*;
use chumsky::error::Simple;
use std::collections::HashMap;

// ---------- Type System ----------

// Type annotations with ownership and reference information
#[derive(Debug, Clone, PartialEq)]
pub enum TypeAnnotation {
    Unknown,                                  // Type not specified
    Integer,                                  // Basic integer type
    Float,                                    // Floating point type
    String,                                   // String type
    Boolean,                                  // Boolean type
    Any,                                      // Any type (dynamic)
    Stack(Box<TypeAnnotation>),               // Stack of elements of a specific type
    Custom(String),                           // User-defined type
    Reference(Box<TypeAnnotation>),           // Reference to a type
    // Ownership-related types for ual 1.5+
    Owned(Box<TypeAnnotation>),               // Owned value
    Borrowed(Box<TypeAnnotation>),            // Borrowed reference
    Mutable(Box<TypeAnnotation>),             // Mutable reference
}

// Stack perspective enum for the perspective system
#[derive(Debug, Clone, PartialEq)]
pub enum StackPerspective {
    LIFO,    // Last In, First Out (traditional stack)
    FIFO,    // First In, First Out (queue)
    MAXFO,   // Maximum First Out (priority queue)
    MINFO,   // Minimum First Out (reverse priority queue)
    Hashed,  // Key-based access (map/dictionary)
}

// ---------- Symbol and Location Information ----------

// Symbol information with enhanced scope tracking
#[derive(Debug, Clone, PartialEq)]
pub struct SymbolInfo {
    pub name: String,
    pub type_annotation: TypeAnnotation,
    pub exported: bool,                  // Whether this symbol is exported (uppercase first letter)
    pub scope_level: usize,              // Scope nesting level (0 = global)
    pub definition_location: Location,   // Where the symbol was defined
    pub references: Vec<Location>,       // Where the symbol is referenced
}

// Location information for diagnostic messages
#[derive(Debug, Clone, PartialEq)]
pub struct Location {
    pub line: usize,
    pub column: usize,
    pub span: std::ops::Range<usize>,    // Character span in source
}

// ---------- Program Structure ----------

// Program with package, imports, and declarations
#[derive(Debug, Clone, PartialEq)]
pub struct Program {
    pub package: PackageDecl,
    pub imports: Vec<ImportDecl>,
    pub decls: Vec<Decl>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct PackageDecl {
    pub name: String,
    pub location: Location,
}

#[derive(Debug, Clone, PartialEq)]
pub struct ImportDecl {
    pub path: String,
    pub location: Location,
}

// ---------- Declarations ----------

#[derive(Debug, Clone, PartialEq)]
pub enum Decl {
    Function(FunctionDecl),
    GlobalVar(GlobalVarDecl),
    Enum(EnumDecl),         // Support for enum declarations
    Constant(ConstantDecl), // Support for constants
}

#[derive(Debug, Clone, PartialEq)]
pub struct FunctionDecl {
    pub name: String,
    pub params: Vec<Parameter>,
    pub return_type: Option<TypeAnnotation>,
    pub body: Vec<Stmt>,
    pub location: Location,
    pub symbol_info: Option<SymbolInfo>,
    pub has_error_handling: bool,    // For @error > annotation
}

#[derive(Debug, Clone, PartialEq)]
pub struct Parameter {
    pub name: String,
    pub type_annotation: Option<TypeAnnotation>,
    pub location: Location,
}

#[derive(Debug, Clone, PartialEq)]
pub struct GlobalVarDecl {
    pub name: String,
    pub expr: Expr,
    pub type_annotation: Option<TypeAnnotation>,
    pub location: Location,
    pub symbol_info: Option<SymbolInfo>,
}

// Enum declarations (ual 1.6 proposal)
#[derive(Debug, Clone, PartialEq)]
pub struct EnumDecl {
    pub name: String,
    pub variants: Vec<EnumVariant>,
    pub location: Location,
    pub symbol_info: Option<SymbolInfo>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct EnumVariant {
    pub name: String,
    pub value: Option<Expr>,  // Optional explicit value
    pub location: Location,
}

// Constants (immutable globals)
#[derive(Debug, Clone, PartialEq)]
pub struct ConstantDecl {
    pub name: String,
    pub expr: Expr,
    pub type_annotation: Option<TypeAnnotation>,
    pub location: Location,
    pub symbol_info: Option<SymbolInfo>,
}

// ---------- Statements ----------

#[derive(Debug, Clone, PartialEq)]
pub enum Stmt {
    Return(Option<Expr>, Location),
    Expr(Expr, Location),
    LocalVar(LocalVarDecl),
    Assign(Vec<LValue>, Vec<Expr>, Location),
    IfTrue { cond: Expr, block: Vec<Stmt>, location: Location },
    IfFalse { cond: Expr, block: Vec<Stmt>, location: Location },
    WhileTrue { cond: Expr, block: Vec<Stmt>, location: Location },
    ForNum { var: String, start: Expr, end: Expr, step: Option<Expr>, block: Vec<Stmt>, location: Location },
    ForGen { var: String, expr: Expr, block: Vec<Stmt>, location: Location },
    Switch { expr: Expr, cases: Vec<Case>, default: Option<Vec<Stmt>>, location: Location },
    StackedMode(StackedModeStmt),
    DeferOp { block: Vec<Stmt>, location: Location },  // Defer block for resource management
    Scope { block: Vec<Stmt>, location: Location },    // Explicit scope block
    // Stack borrowing and segment access
    Borrow { target: LValue, source: StackSegment, mutable: bool, location: Location },
}

// Local variable declaration
#[derive(Debug, Clone, PartialEq)]
pub struct LocalVarDecl {
    pub name: String,
    pub expr: Option<Expr>,
    pub type_annotation: Option<TypeAnnotation>,
    pub location: Location,
    pub symbol_info: Option<SymbolInfo>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct Case {
    pub values: Vec<Expr>,
    pub block: Vec<Stmt>,
    pub location: Location,
}

// L-Value (addressable expression)
#[derive(Debug, Clone, PartialEq)]
pub enum LValue {
    Ident(String, Location),
    IndexAccess(Box<Expr>, Box<Expr>, Location),
    FieldAccess(Box<Expr>, String, Location),
}

// ---------- Expressions ----------

#[derive(Debug, Clone, PartialEq)]
pub enum Expr {
    Ident(String, Location, Option<SymbolInfo>),
    Number(f64, Location),
    String(String, Location),
    Boolean(bool, Location),
    Nil(Location),
    Unary(String, Box<Expr>, Location),
    Binary(Box<Expr>, String, Box<Expr>, Location),
    Call(Box<Expr>, Vec<Expr>, Location),
    Paren(Box<Expr>, Location),
    // Data constructors:
    Table(Vec<TableField>, Location),
    Array(Vec<Expr>, Location),
    Hash(Vec<(Expr, Expr)>, Location),  // Hash literal with key-value pairs
    // JSON literal
    Json(Box<Expr>, Location),
    // Stack operations:
    StackMethod(Box<Expr>, String, Vec<Expr>, Location),
    StackCreation { args: Vec<Expr>, location: Location },
    // Stack perspective operations
    StackPerspective { stack: Box<Expr>, perspective: StackPerspective, location: Location },
    // Result handling and pattern matching:
    Consider { expr: Box<Expr>, clauses: Vec<PatternClause>, location: Location },
    // Borrowed stack segment
    StackSegment { stack: Box<Expr>, range: (Box<Expr>, Box<Expr>), location: Location },
    // Crosstack (ual 1.8)
    Crosstack { base: Box<Expr>, selector: CrossstackSelector, location: Location },
}

#[derive(Debug, Clone, PartialEq)]
pub struct TableField {
    pub key: Option<Expr>,
    pub value: Expr,
    pub location: Location,
}

// Pattern clauses for the consider statement (ual 1.8)
#[derive(Debug, Clone, PartialEq)]
pub enum PatternClause {
    // Original result handling patterns
    IfOk(Expr, Location),
    IfErr(Expr, Location),
    IfErrMatch(Vec<Expr>, Expr, Location),
    // New generalized pattern matching patterns
    IfEqual(Expr, Expr, Location),      // Value to check against, handler
    IfMatch(Expr, Expr, Location),      // Predicate function, handler
    IfType(TypeAnnotation, Expr, Location), // Type to check against, handler
    IfElse(Expr, Location),              // Default handler
}

// Crosstack selector for orthogonal stack access (ual 1.8)
#[derive(Debug, Clone, PartialEq)]
pub enum CrossstackSelector {
    SingleLevel(Box<Expr>),             // Single level, e.g. ~0
    Range(Box<Expr>, Box<Expr>),        // Range of levels, e.g. [0..3]~
    Levels(Vec<Expr>),                  // Specific levels, e.g. [0,2,5]~
    All,                                // All levels, e.g. ~
}

// ---------- Stack Operations ----------

#[derive(Debug, Clone, PartialEq)]
pub struct StackedModeStmt {
    pub target: Option<String>,
    pub operations: Vec<StackOp>,
    pub location: Location,
}

#[derive(Debug, Clone, PartialEq)]
pub enum StackOp {
    Push(Expr, Location),
    Pop(Location),
    Dup(Location),
    Swap(Location),
    Over(Location),
    Rot(Location),
    Add(Location),
    Sub(Location),
    Mul(Location),
    Div(Location),
    // Enhanced stack operations
    PushLiteral(Expr, Location),       // push:literal syntax
    MethodCall(String, Vec<Expr>, Location),
    Transfer(String, String, Location),  // Stack-to-stack transfer
    Perspective(StackPerspective, Location),  // Change perspective
}

// ---------- Helper types ----------

#[derive(Debug, Clone, PartialEq)]
pub struct StackSegment {
    pub stack: Box<Expr>,
    pub range: (Box<Expr>, Box<Expr>),
    pub location: Location,
}

// ---------- Whitespace and Comment Handling ----------

fn ws<'a>() -> impl Parser<'a, &'a str, (), Simple<&'a str>> {
    let lua_comment = just("--").then(take_until(just('\n'))).padded();
    let cpp_comment = just("//").then(take_until(just('\n'))).padded();
    let c_comment = just("/*").then(take_until(just("*/"))).then_ignore(just("*/")).padded();
    
    text::whitespace()
        .or(lua_comment)
        .or(cpp_comment)
        .or(c_comment)
        .repeated()
        .map(|_| ())
}

// Helper to create location info
fn location_from_span(span: std::ops::Range<usize>, input: &str) -> Location {
    let prefix = &input[..span.start];
    let line = prefix.matches('\n').count() + 1;
    let last_newline = prefix.rfind('\n').map(|i| i + 1).unwrap_or(0);
    let column = span.start - last_newline + 1;
    
    Location {
        line,
        column,
        span,
    }
}

// ---------- Parsers for Package and Imports ----------

fn package_decl<'a>(input: &'a str) -> impl Parser<'a, &'a str, PackageDecl, Simple<&'a str>> {
    just("package")
        .padded_by(ws(), ws())
        .ignore_then(
            text::ident()
                .map_with_span(move |name, span| (name, location_from_span(span, input)))
                .padded_by(ws(), ws())
        )
        .map(|(name, location)| PackageDecl { name, location })
}

fn import_decl<'a>(input: &'a str) -> impl Parser<'a, &'a str, ImportDecl, Simple<&'a str>> {
    just("import")
        .padded_by(ws(), ws())
        .ignore_then(
            string_literal(input)
                .map_with_span(move |path, span| (path, location_from_span(span, input)))
                .padded_by(ws(), ws())
        )
        .map(|(path, location)| ImportDecl { path, location })
}

fn string_literal<'a>(input: &'a str) -> impl Parser<'a, &'a str, String, Simple<&'a str>> {
    let inner = none_of("\"").repeated().collect::<String>();
    just('"').ignore_then(inner).then_ignore(just('"')).padded_by(ws(), ws())
}

// ---------- Type Annotation Parser ----------

fn type_annotation<'a>(input: &'a str) -> impl Parser<'a, &'a str, TypeAnnotation, Simple<&'a str>> {
    recursive(|type_anno| {
        let basic_type = select! {
            "Integer" => TypeAnnotation::Integer,
            "Float" => TypeAnnotation::Float,
            "String" => TypeAnnotation::String,
            "Boolean" => TypeAnnotation::Boolean,
            "Any" => TypeAnnotation::Any,
        };
        
        let stack_type = just("Stack")
            .padded_by(ws(), ws())
            .ignore_then(
                type_anno
                    .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
            )
            .map(|inner_type| TypeAnnotation::Stack(Box::new(inner_type)));
            
        let reference_type = just("&")
            .padded_by(ws(), ws())
            .ignore_then(type_anno)
            .map(|inner_type| TypeAnnotation::Reference(Box::new(inner_type)));
            
        let ownership = select! {
            "Owned" => |t| TypeAnnotation::Owned(Box::new(t)),
            "Borrowed" => |t| TypeAnnotation::Borrowed(Box::new(t)),
            "Mutable" => |t| TypeAnnotation::Mutable(Box::new(t)),
        }
        .padded_by(ws(), ws())
        .then(type_anno)
        .map(|(constructor, t)| constructor(t));
        
        let custom_type = text::ident()
            .map(TypeAnnotation::Custom);
            
        choice((
            basic_type,
            stack_type,
            reference_type,
            ownership,
            custom_type,
        ))
    })
}

// ---------- Top-Level Declaration Parsers ----------

fn function_decl<'a>(input: &'a str) -> impl Parser<'a, &'a str, Decl, Simple<&'a str>> {
    let error_annotation = just("@error").padded_by(ws(), ws()).then_ignore(just(">").padded_by(ws(), ws())).or_not();
    
    error_annotation
        .then(
            just("function")
                .padded_by(ws(), ws())
                .ignore_then(
                    text::ident()
                        .map_with_span(move |name, span| (name, location_from_span(span, input)))
                        .padded_by(ws(), ws())
                )
        )
        .then(
            parameter(input)
                .separated_by(just(',').padded_by(ws(), ws()))
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
                .or_not()
                .map(|opt| opt.unwrap_or_else(Vec::new))
        )
        .then(
            // Optional return type annotation
            just("->")
                .padded_by(ws(), ws())
                .ignore_then(type_annotation(input))
                .or_not()
        )
        .then(
            // Support both block styles: {...} and statement list with "end"
            choice((
                statement(input)
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws())),
                statement(input)
                    .repeated()
                    .then_ignore(just("end").padded_by(ws(), ws()))
            ))
        )
        .map_with_span(move |(((has_error, (name, name_loc)), params), return_type, body), span| {
            let full_location = location_from_span(span, input);
            let is_exported = name.chars().next().map_or(false, |c| c.is_uppercase());
            
            Decl::Function(FunctionDecl {
                name: name.clone(),
                params,
                return_type,
                body,
                location: full_location,
                has_error_handling: has_error.is_some(),
                symbol_info: Some(SymbolInfo {
                    name,
                    type_annotation: return_type.unwrap_or(TypeAnnotation::Unknown),
                    exported: is_exported,
                    scope_level: 0,  // Will be updated during semantic analysis
                    definition_location: name_loc,
                    references: Vec::new(),
                }),
            })
        })
}

fn parameter<'a>(input: &'a str) -> impl Parser<'a, &'a str, Parameter, Simple<&'a str>> {
    text::ident()
        .map_with_span(move |name, span| (name, location_from_span(span, input)))
        .then(
            just(":")
                .padded_by(ws(), ws())
                .ignore_then(type_annotation(input))
                .or_not()
        )
        .map(|((name, location), type_annotation)| Parameter {
            name,
            type_annotation,
            location,
        })
}

fn global_var_decl<'a>(input: &'a str) -> impl Parser<'a, &'a str, Decl, Simple<&'a str>> {
    text::ident()
        .map_with_span(move |name, span| (name, location_from_span(span, input)))
        .padded_by(ws(), ws())
        .then(
            just(":")
                .padded_by(ws(), ws())
                .ignore_then(type_annotation(input))
                .or_not()
        )
        .then_ignore(just('=').padded_by(ws(), ws()))
        .then(expr(input).padded_by(ws(), ws()))
        .map_with_span(move |((name, location), type_annotation, expr), span| {
            let full_location = location_from_span(span, input);
            let is_exported = name.chars().next().map_or(false, |c| c.is_uppercase());
            
            Decl::GlobalVar(GlobalVarDecl {
                name: name.clone(),
                expr,
                type_annotation,
                location: full_location,
                symbol_info: Some(SymbolInfo {
                    name,
                    type_annotation: type_annotation.unwrap_or(TypeAnnotation::Unknown),
                    exported: is_exported,
                    scope_level: 0,  // Will be updated during semantic analysis
                    definition_location: location,
                    references: Vec::new(),
                }),
            })
        })
}

fn enum_decl<'a>(input: &'a str) -> impl Parser<'a, &'a str, Decl, Simple<&'a str>> {
    just("enum")
        .padded_by(ws(), ws())
        .ignore_then(
            text::ident()
                .map_with_span(move |name, span| (name, location_from_span(span, input)))
                .padded_by(ws(), ws())
        )
        .then(
            enum_variant(input)
                .separated_by(just(',').padded_by(ws(), ws()))
                .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
        )
        .map_with_span(move |((name, name_loc), variants), span| {
            let full_location = location_from_span(span, input);
            let is_exported = name.chars().next().map_or(false, |c| c.is_uppercase());
            
            Decl::Enum(EnumDecl {
                name: name.clone(),
                variants,
                location: full_location,
                symbol_info: Some(SymbolInfo {
                    name,
                    type_annotation: TypeAnnotation::Custom("Enum".to_string()),
                    exported: is_exported,
                    scope_level: 0,  // Will be updated during semantic analysis
                    definition_location: name_loc,
                    references: Vec::new(),
                }),
            })
        })
}

fn enum_variant<'a>(input: &'a str) -> impl Parser<'a, &'a str, EnumVariant, Simple<&'a str>> {
    text::ident()
        .map_with_span(move |name, span| (name, location_from_span(span, input)))
        .then(
            just('=')
                .padded_by(ws(), ws())
                .ignore_then(expr(input))
                .or_not()
        )
        .map(|((name, location), value)| EnumVariant {
            name,
            value,
            location,
        })
}

fn constant_decl<'a>(input: &'a str) -> impl Parser<'a, &'a str, Decl, Simple<&'a str>> {
    just("const")
        .padded_by(ws(), ws())
        .ignore_then(
            text::ident()
                .map_with_span(move |name, span| (name, location_from_span(span, input)))
                .padded_by(ws(), ws())
        )
        .then(
            just(":")
                .padded_by(ws(), ws())
                .ignore_then(type_annotation(input))
                .or_not()
        )
        .then_ignore(just('=').padded_by(ws(), ws()))
        .then(expr(input).padded_by(ws(), ws()))
        .map_with_span(move |((name, location), type_annotation, expr), span| {
            let full_location = location_from_span(span, input);
            let is_exported = name.chars().next().map_or(false, |c| c.is_uppercase());
            
            Decl::Constant(ConstantDecl {
                name: name.clone(),
                expr,
                type_annotation,
                location: full_location,
                symbol_info: Some(SymbolInfo {
                    name,
                    type_annotation: type_annotation.unwrap_or(TypeAnnotation::Unknown),
                    exported: is_exported,
                    scope_level: 0,  // Will be updated during semantic analysis
                    definition_location: location,
                    references: Vec::new(),
                }),
            })
        })
}

fn top_level_decl<'a>(input: &'a str) -> impl Parser<'a, &'a str, Decl, Simple<&'a str>> {
    choice((
        function_decl(input),
        global_var_decl(input),
        enum_decl(input),
        constant_decl(input),
    ))
}

// ---------- Expression Parsers ----------

fn expr<'a>(input: &'a str) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    recursive(|expr| {
        let atom = choice((
            number_expr(input),
            string_lit_expr(input),
            boolean_expr(input),
            nil_expr(input),
            ident_expr(input),
            paren_expr(input, expr.clone()),
            stack_creation_expr(input, expr.clone()),
            json_literal(input, expr.clone()),
            table_constructor(input, expr.clone()),
            array_constructor(input, expr.clone()),
            hash_literal(input, expr.clone()),
        ));
        
        let call = atom.clone()
            .then(
                expr.clone()
                    .separated_by(just(',').padded_by(ws(), ws()))
                    .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
                    .map_with_span(move |args, span| (args, location_from_span(span, input)))
                    .repeated()
            )
            .foldl(|func, (args, loc)| Expr::Call(Box::new(func), args, loc));
        
        let field_access = call.clone()
            .then(
                just('.')
                    .padded_by(ws(), ws())
                    .ignore_then(text::ident())
                    .map_with_span(move |field, span| (field, location_from_span(span, input)))
                    .repeated()
            )
            .foldl(|obj, (field, loc)| {
                Expr::StackMethod(
                    Box::new(obj),
                    field,
                    Vec::new(),
                    loc
                )
            });
        
        let index_access = field_access.clone()
            .then(
                expr.clone()
                    .delimited_by(just('[').padded_by(ws(), ws()), just(']').padded_by(ws(), ws()))
                    .map_with_span(move |index, span| (index, location_from_span(span, input)))
                    .repeated()
            )
            .foldl(|obj, (index, loc)| {
                Expr::Binary(
                    Box::new(obj),
                    "[]".to_string(),
                    Box::new(index),
                    loc
                )
            });
            
        // Crosstack access (ual 1.8)
        let crosstack_access = index_access.clone()
            .then(
                just('~')
                    .padded_by(ws(), ws())
                    .ignore_then(crosstack_selector(input, expr.clone()))
                    .map_with_span(move |selector, span| (selector, location_from_span(span, input)))
                    .or_not()
            )
            .map(|(base, maybe_selector)| {
                if let Some((selector, loc)) = maybe_selector {
                    Expr::Crosstack { 
                        base: Box::new(base),
                        selector,
                        location: loc,
                    }
                } else {
                    base
                }
            });
        
        // Define operator precedence levels
        let unary = choice((
            just('-').to("-".to_string()),
            just('!').to("!".to_string()),
            just('~').to("~".to_string()),
            just('+').to("+".to_string()),
        ))
        .padded_by(ws(), ws())
        .repeated()
        .then(crosstack_access.clone())
        .map_with_span(move |(ops, expr), span| {
            let loc = location_from_span(span, input);
            ops.into_iter().rev().fold(expr, |acc, op| {
                Expr::Unary(op, Box::new(acc), loc.clone())
            })
        });
        
        let product = unary.clone()
            .then(
                choice((
                    just('*').to("*".to_string()),
                    just('/').to("/".to_string()),
                    just('%').to("%".to_string()),
                ))
                .padded_by(ws(), ws())
                .then(unary.clone())
                .map_with_span(move |((op, right)), span| (op, right, location_from_span(span, input)))
                .repeated()
            )
            .foldl(|left, (op, right, loc)| {
                Expr::Binary(Box::new(left), op, Box::new(right), loc)
            });
        
        let sum = product.clone()
            .then(
                choice((
                    just('+').to("+".to_string()),
                    just('-').to("-".to_string()),
                ))
                .padded_by(ws(), ws())
                .then(product.clone())
                .map_with_span(move |((op, right)), span| (op, right, location_from_span(span, input)))
                .repeated()
            )
            .foldl(|left, (op, right, loc)| {
                Expr::Binary(Box::new(left), op, Box::new(right), loc)
            });
       

let shift = sum.clone()
            .then(
                choice((
                    just("<<").to("<<".to_string()),
                    just(">>").to(">>".to_string()),
                ))
                .padded_by(ws(), ws())
                .then(sum.clone())
                .map_with_span(move |((op, right)), span| (op, right, location_from_span(span, input)))
                .repeated()
            )
            .foldl(|left, (op, right, loc)| {
                Expr::Binary(Box::new(left), op, Box::new(right), loc)
            });
        
        let comparison = shift.clone()
            .then(
                choice((
                    just("<=").to("<=".to_string()),
                    just(">=").to(">=".to_string()),
                    just('<').to("<".to_string()),
                    just('>').to(">".to_string()),
                ))
                .padded_by(ws(), ws())
                .then(shift.clone())
                .map_with_span(move |((op, right)), span| (op, right, location_from_span(span, input)))
                .repeated()
            )
            .foldl(|left, (op, right, loc)| {
                Expr::Binary(Box::new(left), op, Box::new(right), loc)
            });
        
        let equality = comparison.clone()
            .then(
                choice((
                    just("==").to("==".to_string()),
                    just("!=").to("!=".to_string()),
                ))
                .padded_by(ws(), ws())
                .then(comparison.clone())
                .map_with_span(move |((op, right)), span| (op, right, location_from_span(span, input)))
                .repeated()
            )
            .foldl(|left, (op, right, loc)| {
                Expr::Binary(Box::new(left), op, Box::new(right), loc)
            });
        
        let bit_and = equality.clone()
            .then(
                just('&')
                .to("&".to_string())
                .padded_by(ws(), ws())
                .then(equality.clone())
                .map_with_span(move |((op, right)), span| (op, right, location_from_span(span, input)))
                .repeated()
            )
            .foldl(|left, (op, right, loc)| {
                Expr::Binary(Box::new(left), op, Box::new(right), loc)
            });
        
        let bit_xor = bit_and.clone()
            .then(
                just('^')
                .to("^".to_string())
                .padded_by(ws(), ws())
                .then(bit_and.clone())
                .map_with_span(move |((op, right)), span| (op, right, location_from_span(span, input)))
                .repeated()
            )
            .foldl(|left, (op, right, loc)| {
                Expr::Binary(Box::new(left), op, Box::new(right), loc)
            });
        
        let bit_or = bit_xor.clone()
            .then(
                just('|')
                .to("|".to_string())
                .padded_by(ws(), ws())
                .then(bit_xor.clone())
                .map_with_span(move |((op, right)), span| (op, right, location_from_span(span, input)))
                .repeated()
            )
            .foldl(|left, (op, right, loc)| {
                Expr::Binary(Box::new(left), op, Box::new(right), loc)
            });
        
        // Pattern matching with .consider (ual 1.8)
        let consider = bit_or.clone()
            .then(
                just('.')
                    .padded_by(ws(), ws())
                    .ignore_then(just("consider"))
                    .padded_by(ws(), ws())
                    .ignore_then(
                        pattern_clause(input, expr.clone())
                            .repeated()
                            .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
                    )
                    .map_with_span(move |clauses, span| (clauses, location_from_span(span, input)))
                    .or_not()
            )
            .map(|(base_expr, maybe_clauses)| {
                if let Some((clauses, location)) = maybe_clauses {
                    Expr::Consider { 
                        expr: Box::new(base_expr), 
                        clauses,
                        location,
                    }
                } else {
                    base_expr
                }
            });

        // Stack perspective operations
        let perspective_op = consider.clone()
            .then(
                just(':')
                    .padded_by(ws(), ws())
                    .ignore_then(
                        choice((
                            just("lifo").to(StackPerspective::LIFO),
                            just("fifo").to(StackPerspective::FIFO),
                            just("maxfo").to(StackPerspective::MAXFO),
                            just("minfo").to(StackPerspective::MINFO),
                            just("hashed").to(StackPerspective::Hashed),
                        ))
                    )
                    .map_with_span(move |perspective, span| (perspective, location_from_span(span, input)))
                    .or_not()
            )
            .map(|(base_expr, maybe_perspective)| {
                if let Some((perspective, location)) = maybe_perspective {
                    Expr::StackPerspective { 
                        stack: Box::new(base_expr),
                        perspective,
                        location,
                    }
                } else {
                    base_expr
                }
            });

        perspective_op
    })
}

// Pattern clauses for the consider statement (ual 1.8)
fn pattern_clause<'a>(input: &'a str, expr_parser: impl Parser<'a, &'a str, Expr, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, PatternClause, Simple<&'a str>> {
    // Original result handling patterns
    let if_ok = just("if_ok")
        .padded_by(ws(), ws())
        .ignore_then(expr_parser.clone())
        .map_with_span(move |expr, span| PatternClause::IfOk(expr, location_from_span(span, input)));
    
    let if_err = just("if_err")
        .padded_by(ws(), ws())
        .ignore_then(expr_parser.clone())
        .map_with_span(move |expr, span| PatternClause::IfErr(expr, location_from_span(span, input)));
    
    let if_err_match = just("if_err")
        .padded_by(ws(), ws())
        .ignore_then(
            expr_parser.clone()
                .separated_by(just(',').padded_by(ws(), ws()))
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
        )
        .then(expr_parser.clone())
        .map_with_span(move |(patterns, handler), span| 
            PatternClause::IfErrMatch(patterns, handler, location_from_span(span, input))
        );
    
    // New generalized pattern matching patterns (ual 1.8)
    let if_equal = just("if_equal")
        .padded_by(ws(), ws())
        .ignore_then(
            expr_parser.clone()
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
        )
        .then(expr_parser.clone())
        .map_with_span(move |(value, handler), span| 
            PatternClause::IfEqual(value, handler, location_from_span(span, input))
        );
        
    let if_match = just("if_match")
        .padded_by(ws(), ws())
        .ignore_then(
            expr_parser.clone()
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
        )
        .then(expr_parser.clone())
        .map_with_span(move |(pred, handler), span| 
            PatternClause::IfMatch(pred, handler, location_from_span(span, input))
        );
        
    let if_type = just("if_type")
        .padded_by(ws(), ws())
        .ignore_then(
            type_annotation(input)
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
        )
        .then(expr_parser.clone())
        .map_with_span(move |(type_anno, handler), span| 
            PatternClause::IfType(type_anno, handler, location_from_span(span, input))
        );
        
    let if_else = just("if_else")
        .padded_by(ws(), ws())
        .ignore_then(expr_parser.clone())
        .map_with_span(move |handler, span| 
            PatternClause::IfElse(handler, location_from_span(span, input))
        );
    
    choice((
        if_equal,
        if_match,
        if_type,
        if_ok,
        if_err,
        if_err_match,
        if_else,
    ))
}

// Crosstack selector for orthogonal stack access (ual 1.8)
fn crosstack_selector<'a>(input: &'a str, expr_parser: impl Parser<'a, &'a str, Expr, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, CrossstackSelector, Simple<&'a str>> {
    let single_level = expr_parser.clone()
        .map(|expr| CrossstackSelector::SingleLevel(Box::new(expr)));
        
    let range = just('[')
        .padded_by(ws(), ws())
        .ignore_then(expr_parser.clone())
        .then_ignore(just("..").padded_by(ws(), ws()))
        .then(expr_parser.clone())
        .then_ignore(just(']').padded_by(ws(), ws()))
        .map(|(start, end)| CrossstackSelector::Range(Box::new(start), Box::new(end)));
        
    let levels = just('[')
        .padded_by(ws(), ws())
        .ignore_then(
            expr_parser.clone()
                .separated_by(just(',').padded_by(ws(), ws()))
        )
        .then_ignore(just(']').padded_by(ws(), ws()))
        .map(|exprs| CrossstackSelector::Levels(exprs));
        
    let all = just("")
        .map(|_| CrossstackSelector::All);
        
    // Try the more specific patterns first
    choice((
        range,
        levels,
        single_level,
        all,
    ))
}

// Additional basic expression parsers
fn number_expr<'a>(input: &'a str) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    let binary = just("0b")
        .or(just("0B"))
        .ignore_then(filter(|c: &char| *c == '0' || *c == '1').repeated().collect::<String>())
        .try_map(|s: String, span| {
            u64::from_str_radix(&s, 2)
                .map(|v| v as f64)
                .map_err(|e| Simple::custom(span, format!("Invalid binary literal: {}", e)))
        });
    
    let hex = just("0x")
        .or(just("0X"))
        .ignore_then(filter(|c: &char| c.is_digit(16)).repeated().collect::<String>())
        .try_map(|s: String, span| {
            u64::from_str_radix(&s, 16)
                .map(|v| v as f64)
                .map_err(|e| Simple::custom(span, format!("Invalid hexadecimal literal: {}", e)))
        });
    
    let decimal = text::int(10)
        .then(just('.').then(text::digits(10)).or_not())
        .collect::<String>()
        .try_map(|s, span| {
            s.parse::<f64>()
                .map_err(|e| Simple::custom(span, format!("Invalid decimal literal: {}", e)))
        });
    
    choice((binary, hex, decimal))
        .map_with_span(move |val, span| Expr::Number(val, location_from_span(span, input)))
}

fn ident_expr<'a>(input: &'a str) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    text::ident()
        .map_with_span(move |name, span| {
            let loc = location_from_span(span, input);
            Expr::Ident(name, loc, None)
        })
}

fn string_lit_expr<'a>(input: &'a str) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    let inner = none_of("\"").repeated().collect::<String>();
    just('"').ignore_then(inner).then_ignore(just('"'))
        .map_with_span(move |s, span| Expr::String(s, location_from_span(span, input)))
}

fn boolean_expr<'a>(input: &'a str) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    (just("true").to(true).or(just("false").to(false)))
        .map_with_span(move |val, span| Expr::Boolean(val, location_from_span(span, input)))
}

fn nil_expr<'a>(input: &'a str) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    just("nil")
        .map_with_span(move |_, span| Expr::Nil(location_from_span(span, input)))
}

fn paren_expr<'a>(input: &'a str, expr_parser: impl Parser<'a, &'a str, Expr, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    expr_parser
        .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
        .map_with_span(move |e, span| Expr::Paren(Box::new(e), location_from_span(span, input)))
}

// Stack creation with type information
fn stack_creation_expr<'a>(input: &'a str, expr_parser: impl Parser<'a, &'a str, Expr, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    just("Stack.new")
        .padded_by(ws(), ws())
        .ignore_then(
            expr_parser
                .separated_by(just(',').padded_by(ws(), ws()))
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
                .or_not()
                .map(|args| args.unwrap_or_else(Vec::new))
        )
        .map_with_span(move |args, span| {
            Expr::StackCreation { 
                args, 
                location: location_from_span(span, input) 
            }
        })
}

// JSON literal support
fn json_literal<'a>(input: &'a str, expr_parser: impl Parser<'a, &'a str, Expr, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    just("json")
        .padded_by(ws(), ws())
        .ignore_then(
            choice((
                table_constructor(input, expr_parser.clone()),
                array_constructor(input, expr_parser.clone()),
            ))
        )
        .map_with_span(move |expr, span| {
            Expr::Json(Box::new(expr), location_from_span(span, input))
        })
}

// Data constructors
fn table_field<'a>(input: &'a str, expr_parser: impl Parser<'a, &'a str, Expr, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, TableField, Simple<&'a str>> {
    let keydef = choice((
        text::ident()
            .map_with_span(move |s, span| Expr::Ident(s, location_from_span(span, input), None))
            .then_ignore(just('=').padded_by(ws(), ws())),
        expr_parser.clone()
            .delimited_by(just('[').padded_by(ws(), ws()), just(']').padded_by(ws(), ws()))
            .then_ignore(just('=').padded_by(ws(), ws())),
    )).or_not();
    
    keydef.then(expr_parser.padded_by(ws(), ws()))
        .map_with_span(move |(key, value), span| {
            TableField { 
                key, 
                value,
                location: location_from_span(span, input),
            }
        })
}

fn table_constructor<'a>(input: &'a str, expr_parser: impl Parser<'a, &'a str, Expr, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    table_field(input, expr_parser.clone())
        .separated_by(just(',').padded_by(ws(), ws()))
        .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
        .map_with_span(move |fields, span| {
            Expr::Table(fields, location_from_span(span, input))
        })
}

fn array_constructor<'a>(input: &'a str, expr_parser: impl Parser<'a, &'a str, Expr, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    expr_parser
        .separated_by(just(',').padded_by(ws(), ws()))
        .delimited_by(just('[').padded_by(ws(), ws()), just(']').padded_by(ws(), ws()))
        .map_with_span(move |values, span| {
            Expr::Array(values, location_from_span(span, input))
        })
}

// Hash literal with tilde separator (ual 1.7+)
fn hash_literal<'a>(input: &'a str, expr_parser: impl Parser<'a, &'a str, Expr, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Expr, Simple<&'a str>> {
    let key_value = expr_parser.clone()
        .then_ignore(just('~').padded_by(ws(), ws()))
        .then(expr_parser.clone());
        
    key_value
        .separated_by(just(',').padded_by(ws(), ws()))
        .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
        .map_with_span(move |pairs, span| {
            Expr::Hash(pairs, location_from_span(span, input))
        })
}

// ---------- Statement Parsers ----------

fn statement<'a>(input: &'a str) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    recursive(|stmt| {
        choice((
            return_stmt(input),
            local_var_stmt(input),
            if_true_stmt(input, stmt.clone()),
            if_false_stmt(input, stmt.clone()),
            while_true_stmt(input, stmt.clone()),
            for_num_stmt(input, stmt.clone()),
            for_gen_stmt(input, stmt.clone()),
            switch_stmt(input, stmt.clone()),
            defer_stmt(input, stmt.clone()),
            scope_stmt(input, stmt.clone()),
            borrow_stmt(input),
            stacked_mode_stmt(input),
            assign_stmt(input),
            expr_stmt(input),
        ))
    })
}

fn return_stmt<'a>(input: &'a str) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    just("return")
        .padded_by(ws(), ws())
        .ignore_then(expr(input).or_not())
        .map_with_span(move |expr_opt, span| {
            Stmt::Return(expr_opt, location_from_span(span, input))
        })
}

fn local_var_stmt<'a>(input: &'a str) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    just("local")
        .padded_by(ws(), ws())
        .ignore_then(
            text::ident()
                .map_with_span(move |name, span| (name, location_from_span(span, input)))
                .padded_by(ws(), ws())
        )
        .then(
            just(":")
                .padded_by(ws(), ws())
                .ignore_then(type_annotation(input))
                .or_not()
        )
        .then(
            just('=')
                .padded_by(ws(), ws())
                .ignore_then(expr(input))
                .or_not()
        )
        .map_with_span(move |(((name, name_loc), type_anno), init_expr), span| {
            let full_loc = location_from_span(span, input);
            Stmt::LocalVar(LocalVarDecl {
                name: name.clone(),
                expr: init_expr,
                type_annotation: type_anno,
                location: full_loc,
                symbol_info: Some(SymbolInfo {
                    name,
                    type_annotation: type_anno.unwrap_or(TypeAnnotation::Unknown),
                    exported: false,
                    scope_level: 0,  // Updated during semantic analysis
                    definition_location: name_loc,
                    references: Vec::new(),
                }),
            })
        })
}

fn expr_stmt<'a>(input: &'a str) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    expr(input)
        .map_with_span(move |expr, span| {
            Stmt::Expr(expr, location_from_span(span, input))
        })
}

fn assign_stmt<'a>(input: &'a str) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    lvalue(input)
        .separated_by(just(',').padded_by(ws(), ws()))
        .then_ignore(just('=').padded_by(ws(), ws()))
        .then(
            expr(input)
                .separated_by(just(',').padded_by(ws(), ws()))
        )
        .map_with_span(move |(lvalues, exprs), span| {
            Stmt::Assign(lvalues, exprs, location_from_span(span, input))
        })
}

// L-value (addressable expression)
fn lvalue<'a>(input: &'a str) -> impl Parser<'a, &'a str, LValue, Simple<&'a str>> {
    recursive(|lvalue| {
        let ident = text::ident()
            .map_with_span(move |name, span| {
                LValue::Ident(name, location_from_span(span, input))
            });
        
        let field_access = ident.clone()
            .then(
                just('.')
                    .padded_by(ws(), ws())
                    .ignore_then(text::ident())
                    .map_with_span(move |field, span| (field, location_from_span(span, input)))
                    .repeated()
            )
            .foldl(|obj, (field, loc)| {
                LValue::FieldAccess(
                    Box::new(Expr::Ident(
                        match &obj {
                            LValue::Ident(name, _) => name.clone(),
                            _ => "".to_string(), // Placeholder
                        },
                        match &obj {
                            LValue::Ident(_, ident_loc) => ident_loc.clone(),
                            _ => loc.clone(), // Placeholder
                        },
                        None
                    )), 
                    field, 
                    loc
                )
            });
        
        let index_access = ident.clone()
            .then(
                expr(input)
                    .delimited_by(just('[').padded_by(ws(), ws()), just(']').padded_by(ws(), ws()))
                    .map_with_span(move |index, span| (index, location_from_span(span, input)))
                    .repeated()
            )
            .foldl(|obj, (index, loc)| {
                LValue::IndexAccess(
                    Box::new(Expr::Ident(
                        match &obj {
                            LValue::Ident(name, _) => name.clone(),
                            _ => "".to_string(), // Placeholder
                        },
                        match &obj {
                            LValue::Ident(_, ident_loc) => ident_loc.clone(),
                            _ => loc.clone(), // Placeholder
                        },
                        None
                    )),
                    Box::new(index),
                    loc
                )
            });
        
        choice((
            index_access,
            field_access,
            ident,
        ))
    })
}

// If statements
fn if_true_stmt<'a>(input: &'a str, stmt_parser: impl Parser<'a, &'a str, Stmt, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    just("if_true")
        .padded_by(ws(), ws())
        .ignore_then(just('(').padded_by(ws(), ws()))
        .ignore_then(expr(input).padded_by(ws(), ws()))
        .then_ignore(just(')').padded_by(ws(), ws()))
        .then(
            choice((
                stmt_parser
                    .clone()
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws())),
                stmt_parser
                    .clone()
                    .repeated()
                    .then_ignore(just("end_if_true").padded_by(ws(), ws()))
            ))
        )
        .map_with_span(move |(cond, block), span| {
            Stmt::IfTrue { 
                cond, 
                block,
                location: location_from_span(span, input),
            }
        })
}

fn if_false_stmt<'a>(input: &'a str, stmt_parser: impl Parser<'a, &'a str, Stmt, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    just("if_false")
        .padded_by(ws(), ws())
        .ignore_then(just('(').padded_by(ws(), ws()))
        .ignore_then(expr(input).padded_by(ws(), ws()))
        .then_ignore(just(')').padded_by(ws(), ws()))
        .then(
            choice((
                stmt_parser
                    .clone()
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws())),
                stmt_parser
                    .clone()
                    .repeated()
                    .then_ignore(just("end_if_false").padded_by(ws(), ws()))
            ))
        )
        .map_with_span(move |(cond, block), span| {
            Stmt::IfFalse { 
                cond, 
                block,
                location: location_from_span(span, input),
            }
        })
}

// While and for loops
fn while_true_stmt<'a>(input: &'a str, stmt_parser: impl Parser<'a, &'a str, Stmt, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    just("while_true")
        .padded_by(ws(), ws())
        .ignore_then(just('(').padded_by(ws(), ws()))
        .ignore_then(expr(input).padded_by(ws(), ws()))
        .then_ignore(just(')').padded_by(ws(), ws()))
        .then(
            choice((
                stmt_parser
                    .clone()
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws())),
                stmt_parser
                    .clone()
                    .repeated()
                    .then_ignore(just("end_while_true").padded_by(ws(), ws()))
            ))
        )
        .map_with_span(move |(cond, block), span| {
            Stmt::WhileTrue { 
                cond, 
                block,
                location: location_from_span(span, input),
            }
        })
}


fn for_num_stmt<'a>(input: &'a str, stmt_parser: impl Parser<'a, &'a str, Stmt, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    just("for")
        .padded_by(ws(), ws())
        .ignore_then(text::ident().padded_by(ws(), ws()))
        .then_ignore(just('=').padded_by(ws(), ws()))
        .then(expr(input).padded_by(ws(), ws()))
        .then_ignore(just(',').padded_by(ws(), ws()))
        .then(expr(input).padded_by(ws(), ws()))
        .then(
            just(',')
                .padded_by(ws(), ws())
                .ignore_then(expr(input))
                .or_not()
        )
        .then_ignore(just("do").padded_by(ws(), ws()))
        .then(
            choice((
                stmt_parser
                    .clone()
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws())),
                stmt_parser
                    .clone()
                    .repeated()
                    .then_ignore(just("end").padded_by(ws(), ws()))
            ))
        )
        .map_with_span(move |(((((var, start), end), step), block), span| {
            Stmt::ForNum { 
                var, 
                start, 
                end, 
                step,
                block, 
                location: location_from_span(span, input),
            }
        })
}

fn for_gen_stmt<'a>(input: &'a str, stmt_parser: impl Parser<'a, &'a str, Stmt, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    just("for")
        .padded_by(ws(), ws())
        .ignore_then(text::ident().padded_by(ws(), ws()))
        .then_ignore(just("in").padded_by(ws(), ws()))
        .then(expr(input).padded_by(ws(), ws()))
        .then_ignore(just("do").padded_by(ws(), ws()))
        .then(
            choice((
                stmt_parser
                    .clone()
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws())),
                stmt_parser
                    .clone()
                    .repeated()
                    .then_ignore(just("end").padded_by(ws(), ws()))
            ))
        )
        .map_with_span(move |(((var, expr_val), block), span| {
            Stmt::ForGen { 
                var, 
                expr: expr_val, 
                block,
                location: location_from_span(span, input),
            }
        })
}

// Switch statement
fn case_stmt<'a>(input: &'a str, stmt_parser: impl Parser<'a, &'a str, Stmt, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Case, Simple<&'a str>> {
    just("case")
        .padded_by(ws(), ws())
        .ignore_then(
            choice((
                // Single value case
                expr(input),
                // Multiple values in an array (bitmap matching)
                expr(input)
                    .separated_by(just(',').padded_by(ws(), ws()))
                    .delimited_by(just('[').padded_by(ws(), ws()), just(']').padded_by(ws(), ws()))
                    .map(|exprs| {
                        // Create an array expression
                        Expr::Array(exprs, Location { line: 0, column: 0, span: 0..0 })
                    })
            ))
            .map(|expr| match expr {
                Expr::Array(values, _) => values,
                other => vec![other]
            })
        )
        .then_ignore(just(':').padded_by(ws(), ws()))
        .then(
            stmt_parser
                .repeated()
                .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
        )
        .map_with_span(move |(values, block), span| {
            Case { 
                values, 
                block,
                location: location_from_span(span, input),
            }
        })
}

fn switch_stmt<'a>(input: &'a str, stmt_parser: impl Parser<'a, &'a str, Stmt, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    just("switch_case")
        .padded_by(ws(), ws())
        .ignore_then(just('(').padded_by(ws(), ws()))
        .ignore_then(expr(input).padded_by(ws(), ws()))
        .then_ignore(just(')').padded_by(ws(), ws()))
        .then(
            case_stmt(input, stmt_parser.clone())
                .repeated()
                .then(
                    just("default:")
                        .padded_by(ws(), ws())
                        .ignore_then(
                            stmt_parser
                                .repeated()
                                .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
                        )
                        .or_not()
                )
        )
        .then_ignore(just("end_switch").padded_by(ws(), ws()))
        .map_with_span(move |(expr_val, (cases, default)), span| {
            Stmt::Switch { 
                expr: expr_val, 
                cases, 
                default,
                location: location_from_span(span, input),
            }
        })
}

// Defer statement (ual 1.5 proposal)
fn defer_stmt<'a>(input: &'a str, stmt_parser: impl Parser<'a, &'a str, Stmt, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    choice((
        just("defer_op")
            .padded_by(ws(), ws())
            .ignore_then(
                stmt_parser
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
            ),
        just("@defer")
            .padded_by(ws(), ws())
            .then_ignore(just(':').padded_by(ws(), ws()))
            .ignore_then(just("push").padded_by(ws(), ws()))
            .ignore_then(
                stmt_parser
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
            )
    ))
    .map_with_span(move |block, span| {
        Stmt::DeferOp { 
            block,
            location: location_from_span(span, input),
        }
    })
}

// Explicit scope blocks
fn scope_stmt<'a>(input: &'a str, stmt_parser: impl Parser<'a, &'a str, Stmt, Simple<&'a str>> + Clone + 'a) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    just("scope")
        .padded_by(ws(), ws())
        .ignore_then(
            stmt_parser
                .repeated()
                .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
        )
        .map_with_span(move |block, span| {
            Stmt::Scope { 
                block,
                location: location_from_span(span, input),
            }
        })
}

// Stack borrowing and segment access
fn borrow_stmt<'a>(input: &'a str) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    let target = lvalue(input);
    
    let segment = just("borrow")
        .padded_by(ws(), ws())
        .ignore_then(just('(').padded_by(ws(), ws()))
        .ignore_then(
            just('[')
                .padded_by(ws(), ws())
                .ignore_then(expr(input))
                .then_ignore(just("..").padded_by(ws(), ws()))
                .then(expr(input))
                .then_ignore(just(']').padded_by(ws(), ws()))
                .then_ignore(just('@').padded_by(ws(), ws()))
                .then(text::ident())
                .map(|((start, end), stack_name)| {
                    StackSegment {
                        stack: Box::new(Expr::Ident(stack_name, Location { line: 0, column: 0, span: 0..0 }, None)),
                        range: (Box::new(start), Box::new(end)),
                        location: Location { line: 0, column: 0, span: 0..0 }, // Placeholder
                    }
                })
        )
        .then_ignore(just(')').padded_by(ws(), ws()));
        
    let mutable_segment = just("borrow_mut")
        .padded_by(ws(), ws())
        .ignore_then(just('(').padded_by(ws(), ws()))
        .ignore_then(
            just('[')
                .padded_by(ws(), ws())
                .ignore_then(expr(input))
                .then_ignore(just("..").padded_by(ws(), ws()))
                .then(expr(input))
                .then_ignore(just(']').padded_by(ws(), ws()))
                .then_ignore(just('@').padded_by(ws(), ws()))
                .then(text::ident())
                .map(|((start, end), stack_name)| {
                    StackSegment {
                        stack: Box::new(Expr::Ident(stack_name, Location { line: 0, column: 0, span: 0..0 }, None)),
                        range: (Box::new(start), Box::new(end)),
                        location: Location { line: 0, column: 0, span: 0..0 }, // Placeholder
                    }
                })
        )
        .then_ignore(just(')').padded_by(ws(), ws()));
        
    choice((
        // Regular borrow
        target.clone()
            .then_ignore(just('=').padded_by(ws(), ws()))
            .then(segment.clone())
            .map_with_span(move |(target, segment), span| {
                Stmt::Borrow { 
                    target, 
                    source: segment, 
                    mutable: false,
                    location: location_from_span(span, input),
                }
            }),
        // Mutable borrow
        target
            .then_ignore(just('=').padded_by(ws(), ws()))
            .then(mutable_segment)
            .map_with_span(move |(target, segment), span| {
                Stmt::Borrow { 
                    target, 
                    source: segment, 
                    mutable: true,
                    location: location_from_span(span, input),
                }
            }),
        // Shorthand notations (ual 1.6+)
        lvalue(input)
            .then_ignore(just("<<").padded_by(ws(), ws()))
            .then(
                just('[')
                    .padded_by(ws(), ws())
                    .ignore_then(expr(input))
                    .then_ignore(just("..").padded_by(ws(), ws()))
                    .then(expr(input))
                    .then_ignore(just(']').padded_by(ws(), ws()))
                    .then(text::ident())
            )
            .map_with_span(move |(target, ((start, end), stack_name)), span| {
                let loc = location_from_span(span, input);
                Stmt::Borrow { 
                    target, 
                    source: StackSegment {
                        stack: Box::new(Expr::Ident(stack_name, loc.clone(), None)),
                        range: (Box::new(start), Box::new(end)),
                        location: loc.clone(),
                    }, 
                    mutable: false,
                    location: loc,
                }
            }),
        lvalue(input)
            .then_ignore(just("<:mut").padded_by(ws(), ws()))
            .then(
                just('[')
                    .padded_by(ws(), ws())
                    .ignore_then(expr(input))
                    .then_ignore(just("..").padded_by(ws(), ws()))
                    .then(expr(input))
                    .then_ignore(just(']').padded_by(ws(), ws()))
                    .then(text::ident())
            )
            .map_with_span(move |(target, ((start, end), stack_name)), span| {
                let loc = location_from_span(span, input);
                Stmt::Borrow { 
                    target, 
                    source: StackSegment {
                        stack: Box::new(Expr::Ident(stack_name, loc.clone(), None)),
                        range: (Box::new(start), Box::new(end)),
                        location: loc.clone(),
                    }, 
                    mutable: true,
                    location: loc,
                }
            }),
    ))
}

// Stack operations and Stacked Mode
fn stack_op<'a>(input: &'a str) -> impl Parser<'a, &'a str, StackOp, Simple<&'a str>> {
    let push = just("push")
        .padded_by(ws(), ws())
        .then(
            choice((
                // push:literal syntax
                just(':')
                    .padded_by(ws(), ws())
                    .ignore_then(expr(input).padded_by(ws(), ws())),
                // Traditional push(expr) syntax
                expr(input)
                    .separated_by(just(',').padded_by(ws(), ws()))
                    .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
                    .map(|args| {
                        if args.len() == 1 {
                            args[0].clone()
                        } else {
                            // Create tuple expression for multiple args
                            Expr::Array(args, Location { line: 0, column: 0, span: 0..0 })
                        }
                    })
            ))
        )
        .map_with_span(move |(_, expr), span| {
            StackOp::Push(expr, location_from_span(span, input))
        });
    
    // Common stack operations
    let simple_ops = choice((
        just("pop").map_with_span(move |_, span| StackOp::Pop(location_from_span(span, input))),
        just("dup").map_with_span(move |_, span| StackOp::Dup(location_from_span(span, input))),
        just("swap").map_with_span(move |_, span| StackOp::Swap(location_from_span(span, input))),
        just("over").map_with_span(move |_, span| StackOp::Over(location_from_span(span, input))),
        just("rot").map_with_span(move |_, span| StackOp::Rot(location_from_span(span, input))),
        just("add").map_with_span(move |_, span| StackOp::Add(location_from_span(span, input))),
        just("sub").map_with_span(move |_, span| StackOp::Sub(location_from_span(span, input))),
        just("mul").map_with_span(move |_, span| StackOp::Mul(location_from_span(span, input))),
        just("div").map_with_span(move |_, span| StackOp::Div(location_from_span(span, input))),
    ));
        
    let push_literal = text::ident()
        .then(
            just(':')
                .padded_by(ws(), ws())
                .ignore_then(expr(input).padded_by(ws(), ws()))
        )
        .map_with_span(move |(name, expr), span| {
            if name == "push" {
                StackOp::PushLiteral(expr, location_from_span(span, input))
            } else {
                StackOp::MethodCall(name, vec![expr], location_from_span(span, input))
            }
        });
        
    let method_call = text::ident()
        .then(
            expr(input)
                .separated_by(just(',').padded_by(ws(), ws()))
                .delimited_by(just('(').padded_by(ws(), ws()), just(')').padded_by(ws(), ws()))
                .or_not()
                .map(|opt| opt.unwrap_or_else(Vec::new))
        )
        .map_with_span(move |(name, args), span| {
            StackOp::MethodCall(name, args, location_from_span(span, input))
        });
        
    let perspective_op = choice((
        just("lifo").to(StackPerspective::LIFO),
        just("fifo").to(StackPerspective::FIFO),
        just("maxfo").to(StackPerspective::MAXFO),
        just("minfo").to(StackPerspective::MINFO),
        just("hashed").to(StackPerspective::Hashed),
        just("flip").to(StackPerspective::LIFO), // Simplified for this implementation
    ))
    .map_with_span(move |perspective, span| {
        StackOp::Perspective(perspective, location_from_span(span, input))
    });
    
    // Stack transfer operations (borrowing)
    let transfer_op = just("<")
        .padded_by(ws(), ws())
        .ignore_then(text::ident())
        .map_with_span(move |target, span| {
            StackOp::Transfer("move".to_string(), target, location_from_span(span, input))
        });
        
    choice((
        push,
        simple_ops,
        push_literal,
        method_call,
        perspective_op,
        transfer_op,
    ))
}

fn stacked_mode_stmt<'a>(input: &'a str) -> impl Parser<'a, &'a str, Stmt, Simple<&'a str>> {
    // Stack selector with context block
    let selector_with_block = choice((
        // @stack: { operations }
        just('@')
            .padded_by(ws(), ws())
            .ignore_then(text::ident().padded_by(ws(), ws()))
            .then_ignore(just(':').padded_by(ws(), ws()))
            .then(
                stack_op(input)
                    .padded_by(ws(), ws())
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
            ),
        // : { operations } (default data stack)
        just(':')
            .padded_by(ws(), ws())
            .ignore_then(
                stack_op(input)
                    .padded_by(ws(), ws())
                    .repeated()
                    .delimited_by(just('{').padded_by(ws(), ws()), just('}').padded_by(ws(), ws()))
            )
            .map(|ops| (None, ops))
    ));
    
    // Stack selector with operations on a single line
    let selector_with_ops = choice((
        // @stack: operations
        just('@')
            .padded_by(ws(), ws())
            .ignore_then(text::ident().padded_by(ws(), ws()))
            .then_ignore(choice((just(':'), just('>'))).padded_by(ws(), ws()))
            .then(stack_op(input).padded_by(ws(), ws()).repeated()),
        // : operations (default data stack)
        just(':')
            .padded_by(ws(), ws())
            .ignore_then(stack_op(input).padded_by(ws(), ws()).repeated())
            .map(|ops| (None, ops)),
        // > operations (deprecated default data stack syntax)
        just('>')
            .padded_by(ws(), ws())
            .ignore_then(stack_op(input).padded_by(ws(), ws()).repeated())
            .map(|ops| (None, ops))
    ));
    
    // Handle multi-stack operations with semicolons
    let multi_stack_ops = selector_with_ops.clone()
        .then(
            just(';')
                .padded_by(ws(), ws())
                .ignore_then(selector_with_ops)
                .repeated()
        )
        .map(|((first_target, first_ops), rest)| {
            // Convert to a sequence of stacked mode statements
            let mut result = vec![StackedModeStmt {
                target: first_target,
                operations: first_ops,
                location: Location { line: 0, column: 0, span: 0..0 } // Placeholder
            }];
            
            for (target, ops) in rest {
                result.push(StackedModeStmt {
                    target,
                    operations: ops,
                    location: Location { line: 0, column: 0, span: 0..0 } // Placeholder
                });
            }
            
            result
        });
    
    choice((
        // Context block
        selector_with_block
            .map_with_span(move |(target, operations), span| {
                Stmt::StackedMode(StackedModeStmt { 
                    target, 
                    operations,
                    location: location_from_span(span, input),
                })
            }),
        // Multi-stack operations with semicolons
        multi_stack_ops
            .map_with_span(move |stmts, span| {
                // If only one statement, return it directly
                if stmts.len() == 1 {
                    let mut stmt = stmts[0].clone();
                    stmt.location = location_from_span(span, input);
                    Stmt::StackedMode(stmt)
                } else {
                    // Create a scope with multiple stacked mode statements
                    Stmt::Scope {
                        block: stmts.into_iter().map(|s| {
                            Stmt::StackedMode(s)
                        }).collect(),
                        location: location_from_span(span, input),
                    }
                }
            }),
        // Single line operations
        selector_with_ops
            .map_with_span(move |(target, operations), span| {
                Stmt::StackedMode(StackedModeStmt { 
                    target, 
                    operations,
                    location: location_from_span(span, input),
                })
            }),
    ))
}

// ---------- Program Parser with Enhanced Error Recovery ----------

fn program<'a>(input: &'a str) -> impl Parser<'a, &'a str, Program, Simple<&'a str>> {
    package_decl(input)
        .then(import_decl(input).repeated())
        .then(top_level_decl(input).repeated())
        .map(|((pkg, imports), decls)| Program {
            package: pkg,
            imports,
            decls,
        })
        // Improved error recovery: skip until significant token
        .recover_with(skip_then_retry_until([
            ';', '\n', '{', '}', '(', ')', '[', ']'
        ].map(just), end()))
}

// ---------- Semantic Analysis ----------

struct SemanticAnalyzer {
    // Symbol tables for different scopes
    global_symbols: HashMap<String, SymbolInfo>,
    scope_symbols: Vec<HashMap<String, SymbolInfo>>,
    current_scope_level: usize,
}

impl SemanticAnalyzer {
    fn new() -> Self {
        SemanticAnalyzer {
            global_symbols: HashMap::new(),
            scope_symbols: vec![HashMap::new()], // Start with the global scope
            current_scope_level: 0,
        }
    }
    
    fn enter_scope(&mut self) {
        self.scope_symbols.push(HashMap::new());
        self.current_scope_level += 1;
    }
    
    fn exit_scope(&mut self) {
        if self.current_scope_level > 0 {
            self.scope_symbols.pop();
            self.current_scope_level -= 1;
        }
    }
    
    fn add_symbol(&mut self, name: String, mut symbol_info: SymbolInfo) {
        // Update scope level
        symbol_info.scope_level = self.current_scope_level;
        
        if self.current_scope_level == 0 {
            // Global scope
            self.global_symbols.insert(name, symbol_info);
        } else {
            // Local scope
            if let Some(scope) = self.scope_symbols.last_mut() {
                scope.insert(name, symbol_info);
            }
        }
    }
    
    fn lookup_symbol(&self, name: &str) -> Option<&SymbolInfo> {
        // Check local scopes first, from innermost to outermost
        for scope in self.scope_symbols.iter().rev() {
            if let Some(info) = scope.get(name) {
                return Some(info);
            }
        }
        
        // Then check global scope
        self.global_symbols.get(name)
    }
    
    // Process the program and enrich it with semantic information
    fn analyze(&mut self, program: Program) -> Program {
        // Process declarations to build symbol tables
        for decl in &program.decls {
            match decl {
                Decl::Function(func) => {
                    let mut symbol_info = func.symbol_info.clone().unwrap_or_else(|| {
                        SymbolInfo {
                            name: func.name.clone(),
                            type_annotation: TypeAnnotation::Unknown,
                            exported: func.name.chars().next().map_or(false, |c| c.is_uppercase()),
                            scope_level: self.current_scope_level,
                            definition_location: func.location.clone(),
                            references: Vec::new(),
                        }
                    });
                    
                    // Add to symbol table
                    self.add_symbol(func.name.clone(), symbol_info);
                    
                    // Enter function scope
                    self.enter_scope();
                    
                    // Add parameters to function scope
                    for param in &func.params {
                        let param_symbol = SymbolInfo {
                            name: param.name.clone(),
                            type_annotation: param.type_annotation.clone().unwrap_or(TypeAnnotation::Unknown),
                            exported: false,
                            scope_level: self.current_scope_level,
                            definition_location: param.location.clone(),
                            references: Vec::new(),
                        };
                        
                        self.add_symbol(param.name.clone(), param_symbol);
                    }
                    
                    // Process function body
                    // TODO: Walk through statements and enrich with symbol info
                    
                    // Exit function scope
                    self.exit_scope();
                }
                Decl::GlobalVar(var) => {
                    let symbol_info = var.symbol_info.clone().unwrap_or_else(|| {
                        SymbolInfo {
                            name: var.name.clone(),
                            type_annotation: var.type_annotation.clone().unwrap_or(TypeAnnotation::Unknown),
                            exported: var.name.chars().next().map_or(false, |c| c.is_uppercase()),
                            scope_level: self.current_scope_level,
                            definition_location: var.location.clone(),
                            references: Vec::new(),
                        }
                    });
                    
                    // Add to symbol table
                    self.add_symbol(var.name.clone(), symbol_info);
                }
                Decl::Enum(enum_decl) => {
                    let symbol_info = enum_decl.symbol_info.clone().unwrap_or_else(|| {
                        SymbolInfo {
                            name: enum_decl.name.clone(),
                            type_annotation: TypeAnnotation::Custom("Enum".to_string()),
                            exported: enum_decl.name.chars().next().map_or(false, |c| c.is_uppercase()),
                            scope_level: self.current_scope_level,
                            definition_location: enum_decl.location.clone(),
                            references: Vec::new(),
                        }
                    });
                    
                    // Add to symbol table
                    self.add_symbol(enum_decl.name.clone(), symbol_info);
                    
                    // Add enum variants to symbol table as well
                    for variant in &enum_decl.variants {
                        let variant_symbol = SymbolInfo {
                            name: format!("{}.{}", enum_decl.name, variant.name),
                            type_annotation: TypeAnnotation::Custom(enum_decl.name.clone()),
                            exported: enum_decl.name.chars().next().map_or(false, |c| c.is_uppercase()),
                            scope_level: self.current_scope_level,
                            definition_location: variant.location.clone(),
                            references: Vec::new(),
                        };
                        
                        self.add_symbol(format!("{}.{}", enum_decl.name, variant.name), variant_symbol);
                    }
                }
                Decl::Constant(const_decl) => {
                    let symbol_info = const_decl.symbol_info.clone().unwrap_or_else(|| {
                        SymbolInfo {
                            name: const_decl.name.clone(),
                            type_annotation: const_decl.type_annotation.clone().unwrap_or(TypeAnnotation::Unknown),
                            exported: const_decl.name.chars().next().map_or(false, |c| c.is_uppercase()),
                            scope_level: self.current_scope_level,
                            definition_location: const_decl.location.clone(),
                            references: Vec::new(),
                        }
                    });
                    
                    // Add to symbol table
                    self.add_symbol(const_decl.name.clone(), symbol_info);
                }
            }
        }
        
        // Return the enriched program
        program
    }
}


fn semantic_analysis(program: Program) -> Program {
    let mut analyzer = SemanticAnalyzer::new();
    analyzer.analyze(program)
}

// ---------- Main Parser Function ----------

pub fn parse_ual(input: &str) -> Result<Program, Vec<Simple<&str>>> {
    program(input).then_ignore(end()).parse(input)
}

// ---------- Main Entry Point ----------

fn main() {
    let source = r#"
        package Main
        import "fmt"
        import "con"

        /* Function to compute Fibonacci numbers */
        function Fibonacci(n) {
            if_true(n == 0) { return 1 } 
            if_true(n == 1) { return 1 }
            return Fibonacci(n - 1) + Fibonacci(n - 2)
        }

        // Enhanced result handling with pattern matching (ual 1.8)
        result = Fibonacci(5).consider { 
            if_ok fmt.Printf("Success: %d", _1) 
            if_err(ErrorType.NotFound, ErrorType.InvalidInput) fmt.Printf("Expected error: %s", _1)
            if_err fmt.Printf("Unexpected error: %s", _1) 
        }

        // Stacked mode examples with new colon syntax:
        @dstack: push:10 dup add

        // Stack creation with type information
        @Stack.new(Integer, Owned): alias:"numbers"
        
        // Stack perspective example
        @numbers: fifo  // Switch to FIFO perspective
        
        // Defer block for resource management
        defer_op {
            cleanup_resources()
        }
        
        // Borrowed stack segment (ual 1.6)
        window = borrow([0..10]@numbers)
        
        // Crosstack access (ual 1.8)
        @matrix~0: sum  // Sum all elements at level 0 across stacks
        
        // New enum declaration (ual 1.6)
        enum Status {
            OK = 200,
            NotFound = 404,
            ServerError = 500
        }
        
        // Hash literal with tilde separator (ual 1.7)
        config = {"host" ~ "localhost", "port" ~ 8080, "debug" ~ true}
        
        // Switch with bitmap matching
        switch_case(status_code) 
            case [Status.OK, Status.Created]: { process_success() }
            case [Status.NotFound, Status.Gone]: { process_not_found() }
            case Status.ServerError: { process_server_error() }
            default: { process_unknown_status() }
        end_switch
        
        // Generalized pattern matching (ual 1.8)
        value.consider {
            if_equal(0) { fmt.Printf("Zero") }
            if_match(function(v) return v > 10 and v < 20 end) { fmt.Printf("Between 10 and 20") }
            if_type(String) { fmt.Printf("Got string: %s", _1) }
            if_else { fmt.Printf("Something else: %v", _1) }
        }
        
        // JSON literal (ual 1.7)
        data = json{
            "status": "success",
            "results": [1, 2, 3],
            "count": results.length()
        }
    "#;

    match parse_ual(source) {
        Ok(prog) => {
            println!("Successfully parsed AST: {:#?}", prog);
            let enriched_prog = semantic_analysis(prog);
            println!("Semantically analyzed AST: {:#?}", enriched_prog);
        }
        Err(errors) => {
            println!("Errors during parsing:");
            for err in errors {
                println!("Error: {}", err);
            }
        }
    }
}