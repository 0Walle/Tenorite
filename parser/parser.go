package parser

import (
	"fmt"
	"strconv"
	"0Walle/Tenorite/token"
)

type Parser struct {
	Tokens []token.Token
	i      int
	Line   int
	Err    error
}

func NewParser(tokens []token.Token) Parser {
	return Parser { tokens, 0, 1, nil }
}

func (p *Parser) Peek() *token.Token {
	return &p.Tokens[p.i]
}

func (p *Parser) Previous() *token.Token {
	p.Line = p.Tokens[p.i-1].Line
	return &p.Tokens[p.i-1]
}

func (p *Parser) IsAtEnd() bool {
	return p.Peek().Kind == token.EOF
}

func (p *Parser) Check(kind token.TokenKind) bool {
	if p.IsAtEnd() { return false }
	return p.Peek().Kind == kind
}

func (p *Parser) Advance() *token.Token {
	if !p.IsAtEnd() { p.i += 1 }
	return p.Previous()
}

func (p *Parser) Consume(kind token.TokenKind, message string) *token.Token {
	if p.Check(kind) {
		return p.Advance()
	}
	p.Err = p.Error(p.Peek(), "Expected "+message)
	return nil
}

func (p *Parser) Error(where *token.Token, message string) error {
	return fmt.Errorf("Line %d near '%s': %s", where.Line, where.Lexeme, message)
}

// ====== Parsing Methods ======

func (p *Parser) ParseUnit() (Unit, error) {
	var unit Unit

	for !p.IsAtEnd() {
		stmt, err := p.ParseStmt()
		if err != nil { return unit, err }

		unit.Contents = append(unit.Contents, stmt)

		if p.Consume(token.TERMINATOR, "'.' at end of statement.") == nil { return unit, p.Err }
	}

	return unit, nil
}

func (p *Parser) ParseStmt() (Stmt, error) {
	var stmt Stmt
	var nonlocal bool

	if p.Check(token.IF) {
		p.Advance()
		cond, err := p.ParseExpr()
		if err != nil { return stmt, err }
		if p.Consume(token.RETURN, "`return´") == nil {
			return stmt, p.Err
		}
		retval, err := p.ParseExpr()
		if err != nil { return stmt, err }
		return ReturnStmt{ cond, retval }, nil
	} else if p.Check(token.NONLOCAL) {
		p.Advance()
		nonlocal = true
	} else if p.Check(token.LOOP) {
		p.Advance()
		return LoopStmt {}, nil
	} else if p.Check(token.TYPE) {
		type_kw := p.Advance()
		
		ns := p.Consume(token.NAME, "Namespace")
		if ns == nil { return nil, p.Err }
		
		return TypeStmt { type_kw.Line, *ns }, nil
	}
	
	expr, err := p.ParseExpr()
	if err != nil { return stmt, err }

	if nonlocal && !p.Check(token.ASSIGN) {
		return stmt, fmt.Errorf("Expected `:=` in nonlocal assignment")
	}

	if p.Check(token.ASSIGN) {
		assignPos := p.Advance()

		if field, ok := expr.(Field); ok {
			expr, err := p.ParseExpr()
			if err != nil { return stmt, err }
			stmt = FieldAssignStmt { field, assignPos.Line, expr }
			return stmt, nil
		}

		name, ok := expr.(Name)
		if !ok {
			return stmt, fmt.Errorf("Invalid assignment, expected identifier")
		}
		expr, err := p.ParseExpr()
		if err != nil { return stmt, err }
		stmt = AssignStmt { nonlocal, name, assignPos.Line, expr }
	} else if p.Check(token.FN) {
		p.Advance()
		ns, ok := expr.(Name)
		if !ok {
			return stmt, fmt.Errorf("Invalid assignment, expected identifier")
		}
		stmt, err := p.ParseMethodStmt()
		stmt.Namespace = ns
		if err != nil { return stmt, err }
		return stmt, nil
	} else {
		stmt = ExprStmt { expr }
	}

	return stmt, nil
}

func (p *Parser) ParseMethodStmt() (MethStmt, error) {
	var meth MethStmt

	recv := p.Consume(token.NAME, "receiver variable")
	if recv == nil { return meth, p.Err }

	if p.Check(token.NAME) {
		meth.Params = UnaryExpr {
			X: Name { *recv },
			Method: Name { *p.Advance() },
		}
	} else if p.Check(token.OPERATOR) {
		op := p.Advance()

		other := p.Consume(token.NAME, "right side of operator")
		if other == nil { return meth, p.Err }

		meth.Params = BinaryExpr {
			X: Name { *recv },
			Op: Binop { *op },
			Y: Name { *other },
		}
	} else {
		params := []KeyValue{}
		for !p.Check(token.LEFT_BLOCK) {
			key := p.Consume(token.KEY, "key")
			if key == nil { return meth, p.Err }
			param := p.Consume(token.NAME, "name")
			if param == nil { return meth, p.Err }
			params = append(params, KeyValue { Key { *key }, Name { *param }, 0 })
		}
		meth.Params = CallExpr {
			Recv: Name { *recv },
			Args: params,
		}
	}

	if p.Consume(token.LEFT_BLOCK, "`{´") == nil { return meth, p.Err }

	for !p.Check(token.RIGHT_BLOCK) {
		if meth.Body != nil {
			if p.Consume(token.TERMINATOR, "`.´ separator.") == nil { return meth, p.Err }
			if p.Check(token.RIGHT_BLOCK) { break }
		}

		stmt, err := p.ParseStmt()
		if err != nil { return meth, err }

		meth.Body = append(meth.Body, stmt)
	}

	p.Advance()

	return meth, nil
}

func (p *Parser) ParseExpr() (Expr, error) {
	expr, err := p.ParseCallExpr()
	if err != nil { return expr, err }

	for p.Check(token.CASCADE) {
		cascadeTk := p.Advance()

		if p.Check(token.NAME) {
			name := p.Advance()

			rank := 0
			if p.Check(token.AT) {
				p.Advance()
				rank = 1
			}

			expr = UnaryExpr {
				expr, rank, Name { *name },
			}
			continue
		}

		var args []KeyValue = nil
		for p.Check(token.KEY) {
			key := p.Advance()

			rank := 0
			if p.Check(token.AT) {
				p.Advance()
				rank = 1
			}

			value, r, err := p.ParseBinExpr()
			if err != nil { return expr, err }
			if r != 0 { return expr, fmt.Errorf("Cannot have rank here") }
			args = append(args, KeyValue { Key { *key }, value, rank })
		}

		if args != nil {
			expr = CallExpr { expr, 0, args }
		} else {
			p.Error(cascadeTk, "Unexpected Token `:>´")
		}
	}

	return expr, nil
}

func (p *Parser) ParseCallExpr() (Expr, error) {
	var expr Expr

	recv, xrank, err := p.ParseBinExpr()
	if err != nil { return expr, err }

	var args []KeyValue = nil
	for p.Check(token.KEY) {
		key := p.Advance()

		rank, err := p.ParseRank()
		if err != nil { return expr, err }

		value, r, err := p.ParseBinExpr()
		if err != nil { return expr, err }
		if r != 0 { return expr, fmt.Errorf("Cannot have rank here") }
		args = append(args, KeyValue { Key { *key }, value, rank })
	}

	if args != nil {
		expr = CallExpr { recv, xrank, args }
	} else {
		if xrank != 0 {
			return expr, fmt.Errorf("Cannot have rank here")
		}
		expr = recv
	}

	return expr, nil
}

func (p *Parser) ParseBinExpr() (Expr, int, error) {
	var expr Expr

	expr, xrank, err := p.ParseUnExpr()
	if err != nil { return expr, xrank, err }
	for {

		if !(p.Check(token.OPERATOR) || p.Check(token.TYPE)) {
			break
		}

		op := p.Advance()

		if op.Kind == token.TYPE {
			// TODO: not ignore rank
			y, _, err := p.ParseUnExpr()
			if err != nil { return expr, xrank, err }
			expr = BinaryExpr {
				expr, 0, Binop { *op }, y, 0,
			}
			continue
		}

		yrank, err := p.ParseRank()
		if err != nil { return expr, xrank, err }

		y, next_rank, err := p.ParseUnExpr()
		if err != nil { return expr, next_rank, err }
		expr = BinaryExpr {
			expr, xrank, Binop { *op }, y, yrank,
		}
		xrank = next_rank
	}

	return expr, xrank, nil
}

func (p *Parser) ParseUnExpr() (Expr, int, error) {
	var expr Expr

	expr, err := p.ParseTerm()
	if err != nil { return expr, 0, err }

	for {
		rank, err := p.ParseRank()
		if err != nil { return expr, 0, err }

		if p.Check(token.NAME) {
			name := p.Advance()

			expr = UnaryExpr {
				expr, rank, Name { *name },
			}
			continue
		}

		if p.Check(token.LEFT_LIST) {
			lbrack := p.Advance()
			index, err := p.ParseExpr()
			if err != nil { return expr, rank, err }
			if p.Consume(token.RIGHT_LIST, "closing bracket.") == nil { return nil, rank, p.Err }
			expr = IndexExpr { expr, lbrack.Line, index, p.Line }
			continue
		}

		return expr, rank, nil
	}
	
	return expr, 0, nil
}

func (p *Parser) ParseTerm() (Expr, error) {
	tk := p.Advance()

	if tk.Kind == token.NUMBER { return BasicLiteral { *tk, tk.Value }, nil }
	if tk.Kind == token.RAW_STRING { return BasicLiteral { *tk, tk.Value }, nil }
	if tk.Kind == token.REGEX { return BasicLiteral { *tk, tk.Value }, nil }
	if tk.Kind == token.SYMBOL { return Symbol { *tk }, nil }
	if tk.Kind == token.NAME { return Name { *tk }, nil }
	if tk.Kind == token.FIELD { return Field { *tk }, nil }

	if tk.Kind == token.STRING_BEGIN {
		lquote := tk.Line
		var parts []Expr
		for !p.Check(token.STRING_END) {

			if p.Check(token.STRING_LITERAL) {
				tk = p.Advance()
				if tk.Value == "" { continue }
				parts = append(parts, BasicLiteral { *tk, tk.Value } )
			}

			if p.Check(token.LEFT_BLOCK) {
				p.Advance()
				expr, err := p.ParseExpr()
				if err != nil { return expr, err }
				parts = append(parts, expr )

				if p.Consume(token.RIGHT_BLOCK, "closing bracket") == nil { return expr, p.Err }
			}
		}
		rquote := p.Advance().Line

		return StringInterpExpr { lquote, parts, rquote }, nil

	}

	if tk.Kind == token.LEFT_PAREN {
		lparen := tk.Line
		expr, err := p.ParseExpr()
		if err != nil { return expr, err }

		if p.Consume(token.RIGHT_PAREN, "closing parenthesis") == nil {
			return nil, p.Err
		}

		return ParenExpr { lparen, expr, p.Line }, nil
	}

	if tk.Kind == token.LEFT_BLOCK {
		lblock := tk

		var params []Name
		if p.Check(token.PIPE) {
			p.Advance()
			for !p.Check(token.PIPE) {
				name := p.Consume(token.NAME, "parameter name.")
				if name == nil { return nil, p.Err }
				params = append(params, Name { *name })
			}
			p.Advance()
		}

		var body Chunk
		for !p.Check(token.RIGHT_BLOCK) {
			if body != nil {
				if p.Consume(token.TERMINATOR, "`.´ separator.") == nil { return nil, p.Err }
				if p.Check(token.RIGHT_BLOCK) { break }
			}

			stmt, err := p.ParseStmt()
			if err != nil { return nil, err }

			body = append(body, stmt)
		}
		rblock := p.Advance().Line

		return FunctionLiteral { lblock.Line, params, body, rblock }, nil
	}

	if tk.Kind == token.LEFT_LIST {
		lbrack := tk.Line
		if p.Check(token.KEY) {
			return nil, p.Error(tk, "Uninplemented")
		}

		var items []Expr
		for !p.Check(token.RIGHT_LIST) {
			if items != nil {
				if p.Consume(token.SEPARATOR, "`,´ separator.") == nil { return nil, p.Err }
				if p.Check(token.RIGHT_LIST) { break }
			}

			expr, err := p.ParseExpr()
			if err != nil { return nil, err }
			items = append(items, expr)
		}

		rbrack := p.Advance().Line

		return ListLiteral { lbrack, items, rbrack }, nil
	}

	if tk.Kind == token.HASH {
		tk := p.Advance()

		if tk.Kind == token.LEFT_LIST {
			lbrack := tk.Line
		
			var items []TableEntry
			for !p.Check(token.RIGHT_LIST) {
				if items != nil {
					if p.Consume(token.TERMINATOR, "`.´ separator.") == nil { return nil, p.Err }
					if p.Check(token.RIGHT_LIST) { break }
				}

				entry, err := p.ParseTableEntry()
				if err != nil { return nil, err }
				items = append(items, entry)
			}

			rbrack := p.Advance().Line

			return TableLiteral { lbrack, items, rbrack }, nil
		}
	}

	return nil, p.Error(tk, "Unexpected Token `"+tk.Lexeme+"´")
}

func (p *Parser) ParseTableEntry() (TableEntry, error) {
	var entry TableEntry

	if p.Check(token.KEY) {
		key := p.Advance()
		entry.Key = Symbol { token.Token { Value: key.Value[:len(key.Value)-1] } }
		expr, err := p.ParseExpr()
		if err != nil { return entry, err }
		entry.Value = expr
		return entry, nil
	}
	if p.Check(token.LEFT_PAREN) {
		p.Advance()
		key, err := p.ParseExpr()

		if p.Consume(token.RIGHT_PAREN, "closing `)´.") == nil { return entry, p.Err }
		if p.Consume(token.COLON, "`:´ in table key.") == nil { return entry, p.Err }

		expr, err := p.ParseExpr()
		if err != nil { return entry, err }
		entry.Key = key
		entry.Value = expr
		return entry, nil
	}

	tk := p.Advance()

	return entry, p.Error(tk, "Unexpected Token `"+tk.Lexeme+"´ in table literal")
}

func (p *Parser) ParseRank() (int, error) {
	rank := 0
	if p.Check(token.AT) {
		p.Advance()
		rank = 1
		if p.Check(token.NUMBER) {
			tk := p.Advance()
			n, err := strconv.ParseInt(tk.Value, 10, 64)
			if err != nil {
				return rank, err
			}
			rank = int(n)
		}
	}

	return rank, nil
}