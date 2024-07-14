package token

import (
    "strings"
    "unicode/utf8"
)

func isDigit(c rune) bool {
    return c >= '0' && c <= '9'
}

func isLowerAlpha(c rune) bool {
    return c >= 'a' && c <= 'z'
}

func isNameFirst(c rune) bool {
    return isLowerAlpha(c) || c == '_' || isUpperAlpha(c)
}

func isName(c rune) bool {
    return isLowerAlpha(c) || isUpperAlpha(c) || isDigit(c) || c == '_'
}

func isUpperAlpha(c rune) bool {
    return c >= 'A' && c <= 'Z'
}

func isOperator(c rune) bool {
    return strings.ContainsRune("\\-+*/^~<=>!;$%?", c)
}

const (
    MAIN_MODE = iota
    STR_MODE
)

type Scanner struct {
    tokens   []Token
    lexeme   strings.Builder
    source   string
    line     int
    start    int
    curr     int
    mode     []int
    reporter ErrorReporter
    hasError bool
}

type ErrorReporter interface {
    Report(line int, where string, message string)
}

var eof = rune(0)

func NewScanner(source string, reporter ErrorReporter) Scanner {
    return Scanner {
        tokens: nil,
        source: source,
        reporter: reporter,
        line: 1,
        start: 0,
        curr: 0,
        mode: []int{MAIN_MODE},
    }
}

func (s *Scanner) Read() rune {
    r, size := utf8.DecodeRuneInString(s.source[s.curr:])
    if size == 0 {
        return eof
    }
    s.lexeme.WriteRune(r)
    s.curr += size
    return r
}

func (s *Scanner) Peek() rune {
    r, size := utf8.DecodeRuneInString(s.source[s.curr:])
    if size == 0 {
        return eof
    }
    return r
}

func (s *Scanner) Peek2() rune {
    _, size1 := utf8.DecodeRuneInString(s.source[s.curr:])
    if size1 == 0 { return eof }
    r, size2 := utf8.DecodeRuneInString(s.source[s.curr+size1:])
    if size2 == 0 { return eof }
    return r
}

func (s *Scanner) Match(expected rune) bool {
    r, size := utf8.DecodeRuneInString(s.source[s.curr:])
    if r != expected {
        return false
    }
    s.lexeme.WriteRune(r)
    s.curr += size
    return true
}

func (s *Scanner) Report(message string) {
    s.hasError = true
    s.reporter.Report(s.line, s.lexeme.String(), message)
}

func (s *Scanner) pushLiteral(kind TokenKind, stringValue string) {
    s.tokens = append(s.tokens, Token {
        Kind: kind,
        Lexeme: s.lexeme.String(),
        Value: stringValue,
        Line: s.line,
    })
}

func (s *Scanner) push(kind TokenKind) {
    s.pushLiteral(kind, "")
}

func (s *Scanner) lastEndsExpr() bool {
    if len(s.tokens) < 1 {
        return false
    }
    tk := s.tokens[len(s.tokens)-1]
    return TokenEndsExpression(tk)
}

func (scanner *Scanner) Scan() ([]Token, bool) {
    for {
        if scanner.Peek() == eof {
            break
        }

        mode := scanner.mode[len(scanner.mode)-1]
        switch mode {
        case MAIN_MODE:
            scanner.lexeme.Reset()
            scanner.scanToken(mode)
        case STR_MODE:
            scanner.scanString()
        }
    }

    scanner.lexeme.Reset()
    scanner.push(EOF)

    return scanner.tokens, scanner.hasError
}

func (scanner *Scanner) scanToken(mode int) {
    tk := scanner.Read()
    switch tk {
    case ' ', '\r', '\t':
    case '\n':
        scanner.line += 1
        if scanner.lastEndsExpr() {
            scanner.push(TERMINATOR)
        }
    case '(': scanner.push(LEFT_PAREN)
    case ')': scanner.push(RIGHT_PAREN)
    case '[': scanner.push(LEFT_LIST)
    case ']': scanner.push(RIGHT_LIST)
    case '|': scanner.push(PIPE)
    case ',': scanner.push(SEPARATOR)
    case '@': scanner.push(AT)
    case '{':
        scanner.mode = append(scanner.mode, MAIN_MODE)
        scanner.push(LEFT_BLOCK)
    case '}':
        scanner.mode = scanner.mode[:len(scanner.mode)-1]
        scanner.push(RIGHT_BLOCK)
        scanner.lexeme.Reset()
    case '&':
        peek_tk := scanner.Peek()
        if isNameFirst(peek_tk) {
            peek_tk = scanner.Read()
            name := scanner.scanName(peek_tk)
            scanner.pushLiteral(FIELD, name)
        } else {
            scanner.Report("Unexpected character.")
            scanner.push(ILEGAL)
        }
    case ':':
        if scanner.Match('=') {
            scanner.push(ASSIGN)
        } else if scanner.Match('>') {

            if scanner.tokens[len(scanner.tokens)-1].Kind == TERMINATOR {
                scanner.tokens = scanner.tokens[:len(scanner.tokens)-1]
            }

            scanner.push(CASCADE)
        } else {
            scanner.push(COLON)
        }
    case '.':
        if scanner.Match('.') {
            for {
                if tk := scanner.Peek(); tk == '\n' || tk == eof {
                    break
                }
                scanner.Read()
            }
        } else {
            scanner.push(TERMINATOR)
        }
    case '"':
        scanner.mode = append(scanner.mode, STR_MODE)
        scanner.push(STRING_BEGIN)
        scanner.lexeme.Reset()
    case '\'':
        str := scanner.scanRawString()
        scanner.pushLiteral(RAW_STRING, str)
    case '#':
        peek_tk := scanner.Read()
        if peek_tk == '\'' {
            str := scanner.scanRawString()
            scanner.pushLiteral(REGEX, str)
            return
        }
        if isOperator(peek_tk) {
            name := scanner.scanOperator(peek_tk)
            scanner.pushLiteral(SYMBOL, name)
            return
        }
        if isNameFirst(peek_tk) {
            name := scanner.scanName(peek_tk)

            if scanner.Match(':') {
                keysel := name+":"
                for isNameFirst(scanner.Peek()) {
                    tk := scanner.Read()
                    name := scanner.scanName(tk)
                    if scanner.Match(':') {
                        keysel += name+":"
                    } else {
                        scanner.Report("Expected `:´ at key.")
                        return
                    }
                }
                
                scanner.pushLiteral(SYMBOL, keysel)
                return
            }

            scanner.pushLiteral(SYMBOL, name)
            return
        }
        scanner.Report("Unexpected character.")
    case '-':
        peek_tk := scanner.Peek()
        if isDigit(peek_tk) {
            scanner.scanNumber('-')
            return
        }
        name := scanner.scanOperator(tk)
        scanner.pushLiteral(OPERATOR, name)
    default:
        if isDigit(tk) {
            scanner.scanNumber(tk)
        } else if isOperator(tk) {
            name := scanner.scanOperator(tk)
            scanner.pushLiteral(OPERATOR, name)
        } else if isNameFirst(tk) {
            name := scanner.scanName(tk)

            if scanner.Match(':') {
                scanner.pushLiteral(KEY, name+":")
                return
            }

            switch name {
            case "fn": scanner.push(FN);
            case "return": scanner.push(RETURN);
            case "nonlocal": scanner.push(NONLOCAL);
            case "loop": scanner.push(LOOP);
            case "if": scanner.push(IF);
            case "import": scanner.push(IMPORT);
            case "type": scanner.push(TYPE);
            default:
                scanner.pushLiteral(NAME, name)
            }
        } else {
            scanner.Report("Unexpected character.")
        }
    }
}

func (scanner *Scanner) scanString() {
    tk := scanner.Peek()
    
    switch tk {
    case '"':
        scanner.pushLiteral(STRING_LITERAL, scanner.lexeme.String())
        scanner.lexeme.Reset()

        scanner.Read()
        scanner.mode = scanner.mode[:len(scanner.mode)-1]
        scanner.push(STRING_END)
    case '\\':
        scanner.pushLiteral(STRING_LITERAL, scanner.lexeme.String())
        scanner.lexeme.Reset()

        scanner.Read()
        tk = scanner.Read()
        switch tk {
            case 'n':
                scanner.pushLiteral(STRING_LITERAL, "\n")
                scanner.lexeme.Reset()
            case 't':
                scanner.pushLiteral(STRING_LITERAL, "\t")
                scanner.lexeme.Reset()
            case 'r':
                scanner.pushLiteral(STRING_LITERAL, "\r")
                scanner.lexeme.Reset()
            case '\\':
                scanner.pushLiteral(STRING_LITERAL, "\\")
                scanner.lexeme.Reset()
            case '"':
                scanner.pushLiteral(STRING_LITERAL, "\"")
                scanner.lexeme.Reset()
            case '#':
                scanner.pushLiteral(STRING_LITERAL, "#")
                scanner.lexeme.Reset()
            case eof:
                scanner.Report("Unterminated string.")
                return
        }
    case '#':
        if scanner.Peek2() == '{' {
            scanner.pushLiteral(STRING_LITERAL, scanner.lexeme.String())
            scanner.lexeme.Reset()

            scanner.Read()
            scanner.Read()
            scanner.mode = append(scanner.mode, MAIN_MODE)
            scanner.push(LEFT_BLOCK)
        } else {
            scanner.Read()
        }
    default:
        scanner.Read()
    }

    return
}

func (scanner *Scanner) scanRawString() string {
    var result strings.Builder

    var tk rune
    for {
        tk = scanner.Read()
        if tk == eof { break }

        if tk != '\'' {
            result.WriteRune(tk)
            continue
        }

        if scanner.Match('\'') {
            result.WriteRune('\'')
            continue
        }

        break
    }

    if tk == eof {
        scanner.Report("Unterminated string.")
        return ""
    }

    return result.String()
}

func (scanner *Scanner) scanInteger(result *strings.Builder) {
    var tk rune
    for {
        tk = scanner.Peek()
        if tk == '\'' {
            scanner.Read()
            continue
        }
        if !isDigit(tk) { break }
        result.WriteRune(tk)
        scanner.Read()
    }
}

func (scanner *Scanner) scanNumber(first rune) {
    var result strings.Builder

    result.WriteRune(first)

    scanner.scanInteger(&result)

    if scanner.Peek() == '.' && isDigit(scanner.Peek2()) {
        scanner.Read()
        result.WriteRune('.')
        scanner.scanInteger(&result)
    }

    scanner.pushLiteral(NUMBER, result.String())
}

func (scanner *Scanner) scanName(first rune) string {
    var result strings.Builder

    result.WriteRune(first)

    for isName(scanner.Peek()) {
        result.WriteRune(scanner.Read())
    }

    return result.String()
}

func (scanner *Scanner) scanOperator(first rune) string {
    var result strings.Builder

    result.WriteRune(first)

    for isOperator(scanner.Peek()) {
        result.WriteRune(scanner.Read())
    }

    return result.String()
}