package parser

import (
    "0Walle/Tenorite/token"
)

/*
chunk ::= stat `.´ [`loop´]

stat ::=
     [`nonlocal´] Name `:=´ exp | 
     `type´ Name `:=´ Name | 
     `type´ Name `fn´ Name params `{´ chunk `}´ | 
     Name `fn´ Name params `{´ chunk `}´ | 
     `if´ exp `return´ exp | 
     exp | 

params ::= Name | {Key Name}

exp ::= binexp [rank] {Key [rank] binexp}
binexp ::= unexp [rank] {Binop [rank] unexp}
unexp ::= term { [rank] (Name | `[´ exp `]´) }

rank ::= `@´ [Number]
    
term ::=
    Number | 
    String | 
    symbol | 
    Name | 
    `(´ exp `)´ | 
    function | 
    listliteral |
    tableliteral

function ::= `{´ `|´ {Name} `|´ block `}´

listliteral ::= `[´ [exp {`,´ exp}] `]´

tableliteral ::= `#´ `[´ [tableentry {`,´ tableentry} ] `]´

tableentry ::=
    `(´ exp `)´ `:´ binexp |
    Key binexp


symbol ::= `#´Name | `#´Binop | `#´Key{Key}

*/

type Printer interface {
    Print(int)
}


type Chunk []Stmt

type Unit struct {
    Contents  Chunk
}

type Key struct {
    Value  token.Token
}

type KeyValue struct {
    Key    Key
    Value  Expr
    Rank   int
}

type TableEntry struct {
    Key    Expr
    Value  Expr
}

type Stmt interface {
    stmtNode()
}

type AssignStmt struct {
    NonLocal   bool
    Name       Name
    AssignPos  int
    Value      Expr
}

type FieldAssignStmt struct {
    Name       Field
    AssignPos  int
    Value      Expr
}

type MethStmt struct {
    Namespace  Name
    FnPos      int
    Params     Expr
    Body       Chunk
}

type TypeStmt struct {
    Type       int
    Namespace  token.Token
}

type ReturnStmt struct {
    Cond    Expr
    Return  Expr
}

type LoopStmt struct { }

type ExprStmt struct {
    X  Expr
}

func (_ AssignStmt) stmtNode()      {}
func (_ FieldAssignStmt) stmtNode() {}
func (_ MethStmt) stmtNode()        {}
func (_ TypeStmt) stmtNode()        {}
func (_ ReturnStmt) stmtNode()      {}
func (_ LoopStmt) stmtNode()        {}
func (_ ExprStmt) stmtNode()        {}


type Expr interface {
    // Print(int)
    Line() int
    exprNode()
}

type CallExpr struct {
    Recv    Expr
    RRank   int
    Args    []KeyValue
}

type StringInterpExpr struct {
    Lquote  int
    Parts   []Expr
    Rquote  int
}

type BinaryExpr struct {
    X       Expr
    XRank   int
    Op      Binop
    Y       Expr
    YRank   int
}

type UnaryExpr struct {
    X       Expr
    XRank   int
    Method  Name
}

type IndexExpr struct {
    X       Expr
    Lbrack  int
    Y       Expr
    Rbrack  int
}

type ParenExpr struct {
    Lparen  int
    X       Expr
    Rparen  int
}

type ListLiteral struct {
    Lbrack  int
    List    []Expr
    Rbrack  int
}

type TableLiteral struct {
    Lbrack  int
    Items   []TableEntry
    Rbrack  int
}

type FunctionLiteral struct {
    Lblock  int
    Params  []Name
    Body    Chunk
    Rblock  int
}

type Name struct {
    Value  token.Token
}

type Field struct {
    Value  token.Token
}

type Binop struct {
    Op     token.Token
}

type Symbol struct {
    Value  token.Token
}

type BasicLiteral struct {
    Kind   token.Token
    Value  string
}

func (_ CallExpr) exprNode()          {}
func (_ BinaryExpr) exprNode()        {}
func (_ UnaryExpr) exprNode()         {}
func (_ IndexExpr) exprNode()         {}
func (_ ParenExpr) exprNode()         {}
func (_ FunctionLiteral) exprNode()   {}
func (_ Name) exprNode()              {}
func (_ Field) exprNode()             {}
func (_ ListLiteral) exprNode()       {}
func (_ TableLiteral) exprNode()      {}
func (_ BasicLiteral) exprNode()      {}
func (_ Binop) exprNode()             {}
func (_ Symbol) exprNode()            {}
func (_ StringInterpExpr) exprNode()  {}

func (expr CallExpr) Line() int          { return expr.Args[0].Key.Value.Line }
func (expr BinaryExpr) Line() int        { return expr.Op.Op.Line }
func (expr UnaryExpr) Line() int         { return expr.Method.Value.Line }
func (expr IndexExpr) Line() int         { return expr.X.Line() }
func (expr ParenExpr) Line() int         { return expr.X.Line() }
func (expr FunctionLiteral) Line() int   { return expr.Lblock }
func (expr Name) Line() int              { return expr.Value.Line }
func (expr Field) Line() int             { return expr.Value.Line }
func (expr ListLiteral) Line() int       { return expr.Lbrack }
func (expr TableLiteral) Line() int      { return expr.Lbrack }
func (expr BasicLiteral) Line() int      { return expr.Kind.Line }
func (expr Binop) Line() int             { return expr.Op.Line }
func (expr Symbol) Line() int            { return expr.Value.Line }
func (expr StringInterpExpr) Line() int  { return expr.Lquote }