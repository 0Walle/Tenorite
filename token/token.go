package token

import ()

type TokenKind int

const (
    ILEGAL TokenKind = iota
    EOF
    COMMENT

    NAME
    FIELD
    KEY
    NUMBER
    SYMBOL
    RAW_STRING
    REGEX

    STRING_BEGIN
    STRING_LITERAL
    STRING_END

    OPERATOR
    ASSIGN
    CASCADE
    HASH
    TERMINATOR
    SEPARATOR
    AT
    COLON
    PIPE
    LEFT_PAREN
    RIGHT_PAREN
    LEFT_BLOCK
    RIGHT_BLOCK
    LEFT_LIST
    RIGHT_LIST

    FN
    RETURN
    NONLOCAL
    LOOP
    IF
    IMPORT
    TYPE
)

type Token struct {
    Kind       TokenKind
    Lexeme     string
    Value      string
    Line       int
}

func TokenEndsExpression(tk Token) bool {
    switch tk.Kind {
    case ASSIGN, LEFT_PAREN, LEFT_BLOCK, LEFT_LIST,
        PIPE, OPERATOR, TERMINATOR, CASCADE, SEPARATOR:
        return false
    }
    return true
}