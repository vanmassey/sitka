package parser

import (
	"fmt"
	"strconv"

	"github.com/vanmassey/sitka/internal/ast"
	"github.com/vanmassey/sitka/internal/lexer"
)

type Parser struct {
	tokens   []lexer.Token
	position int
}

func Parse(tokens []lexer.Token) (ast.Program, error) {
	parser := &Parser{tokens: tokens}
	statements := make([]ast.Stmt, 0)
	parser.skipSeparators()
	for !parser.check(lexer.TokenEnd) {
		statement, err := parser.parseDeclaration()
		if err != nil {
			return ast.Program{}, err
		}
		statements = append(statements, statement)
		parser.skipSeparators()
	}
	return ast.Program{Statements: statements}, nil
}

func ParseExpression(tokens []lexer.Token) (ast.Expr, error) {
	parser := &Parser{tokens: tokens}
	return parser.parseExpression()
}

func (p *Parser) parseDeclaration() (ast.Stmt, error) {
	if p.match(lexer.TokenLet) {
		return p.parseVarDeclaration()
	}
	if p.match(lexer.TokenFun) {
		return p.parseFunctionDeclaration()
	}
	return p.parseStatement()
}

func (p *Parser) parseVarDeclaration() (ast.Stmt, error) {
	name, err := p.consume(lexer.TokenIdentifier, "expected variable name")
	if err != nil {
		return nil, err
	}
	var initializer ast.Expr
	if p.match(lexer.TokenEqual) {
		initializer, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}
	p.consumeOptionalSeparator()
	return ast.VarDecl{Name: name.Text, Initializer: initializer, Line: name.Line}, nil
}

func (p *Parser) parseFunctionDeclaration() (ast.Stmt, error) {
	name, err := p.consume(lexer.TokenIdentifier, "expected function name")
	if err != nil {
		return nil, err
	}
	if _, err := p.consume(lexer.TokenLeftParen, "expected '(' after function name"); err != nil {
		return nil, err
	}
	parameters := make([]string, 0)
	if !p.check(lexer.TokenRightParen) {
		for {
			parameter, err := p.consume(lexer.TokenIdentifier, "expected parameter name")
			if err != nil {
				return nil, err
			}
			parameters = append(parameters, parameter.Text)
			if !p.match(lexer.TokenComma) {
				break
			}
		}
	}
	if _, err := p.consume(lexer.TokenRightParen, "expected ')' after parameters"); err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return ast.FunctionDecl{Name: name.Text, Parameters: parameters, Body: body, Line: name.Line}, nil
}

func (p *Parser) parseStatement() (ast.Stmt, error) {
	if p.match(lexer.TokenPrint) {
		return p.parsePrint()
	}
	if p.match(lexer.TokenIf) {
		return p.parseIf()
	}
	if p.match(lexer.TokenWhile) {
		return p.parseWhile()
	}
	if p.match(lexer.TokenFor) {
		return p.parseFor()
	}
	if p.check(lexer.TokenLeftBrace) && !p.mapLiteralAhead() {
		p.advance()
		block, err := p.parseBlockContents()
		if err != nil {
			return nil, err
		}
		return block, nil
	}
	if p.match(lexer.TokenBreak) {
		token := p.previous()
		p.consumeOptionalSeparator()
		return ast.BreakStmt{Line: token.Line}, nil
	}
	if p.match(lexer.TokenContinue) {
		token := p.previous()
		p.consumeOptionalSeparator()
		return ast.ContinueStmt{Line: token.Line}, nil
	}
	if p.match(lexer.TokenReturn) {
		token := p.previous()
		var value ast.Expr
		if !p.checkSeparator() && !p.check(lexer.TokenRightBrace) && !p.check(lexer.TokenEnd) {
			var err error
			value, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}
		p.consumeOptionalSeparator()
		return ast.ReturnStmt{Value: value, Line: token.Line}, nil
	}
	return p.parseExpressionStatement()
}

func (p *Parser) parsePrint() (ast.Stmt, error) {
	token := p.previous()
	expressions := make([]ast.Expr, 0)
	if !p.checkSeparator() && !p.check(lexer.TokenRightBrace) && !p.check(lexer.TokenEnd) {
		for {
			expression, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			expressions = append(expressions, expression)
			if !p.match(lexer.TokenComma) {
				break
			}
		}
	}
	p.consumeOptionalSeparator()
	return ast.PrintStmt{Expressions: expressions, Line: token.Line}, nil
}

func (p *Parser) parseIf() (ast.Stmt, error) {
	token := p.previous()
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	thenBlock, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	var elseBlock *ast.Block
	if p.match(lexer.TokenElse) {
		if p.check(lexer.TokenIf) {
			p.advance()
			nested, err := p.parseIf()
			if err != nil {
				return nil, err
			}
			nestedBlock := ast.Block{Statements: []ast.Stmt{nested}, Line: nested.(ast.IfStmt).Line}
			elseBlock = &nestedBlock
		} else {
			block, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			elseBlock = &block
		}
	}
	return ast.IfStmt{Condition: condition, Then: thenBlock, Else: elseBlock, Line: token.Line}, nil
}

func (p *Parser) parseWhile() (ast.Stmt, error) {
	token := p.previous()
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return ast.WhileStmt{Condition: condition, Body: body, Line: token.Line}, nil
}

func (p *Parser) parseFor() (ast.Stmt, error) {
	token := p.previous()
	variable, err := p.consume(lexer.TokenIdentifier, "expected variable name after 'for'")
	if err != nil {
		return nil, err
	}
	if _, err := p.consume(lexer.TokenIn, "expected 'in' after for variable"); err != nil {
		return nil, err
	}
	iterable, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return ast.ForStmt{Variable: variable.Text, Iterable: iterable, Body: body, Line: token.Line}, nil
}

func (p *Parser) parseBlock() (ast.Block, error) {
	open, err := p.consume(lexer.TokenLeftBrace, "expected '{'")
	if err != nil {
		return ast.Block{}, err
	}
	block, err := p.parseBlockContents()
	if err != nil {
		return ast.Block{}, err
	}
	block.Line = open.Line
	return block, nil
}

func (p *Parser) parseBlockContents() (ast.Block, error) {
	statements := make([]ast.Stmt, 0)
	p.skipSeparators()
	for !p.check(lexer.TokenRightBrace) && !p.check(lexer.TokenEnd) {
		statement, err := p.parseDeclaration()
		if err != nil {
			return ast.Block{}, err
		}
		statements = append(statements, statement)
		p.skipSeparators()
	}
	if _, err := p.consume(lexer.TokenRightBrace, "expected '}' to close block"); err != nil {
		return ast.Block{}, err
	}
	return ast.Block{Statements: statements}, nil
}

func (p *Parser) parseExpressionStatement() (ast.Stmt, error) {
	expression, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.consumeOptionalSeparator()
	return ast.ExpressionStmt{Expression: expression, Line: p.line()}, nil
}

func (p *Parser) parseExpression() (ast.Expr, error) {
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() (ast.Expr, error) {
	expression, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.match(lexer.TokenEqual) {
		equals := p.previous()
		value, err := p.parseAssignment()
		if err != nil {
			return nil, err
		}
		switch target := expression.(type) {
		case ast.Variable:
			return ast.Assign{Name: target.Name, Value: value, Line: equals.Line}, nil
		case ast.Index:
			return ast.IndexAssign{Object: target.Object, Index: target.Index, Value: value, Line: equals.Line}, nil
		default:
			return nil, fmt.Errorf("invalid assignment target at line %d", equals.Line)
		}
	}
	return expression, nil
}

func (p *Parser) parseOr() (ast.Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenOr) {
		operator := p.previous()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = ast.Logical{Left: left, Operator: operator.Text, Right: right, Line: operator.Line}
	}
	return left, nil
}

func (p *Parser) parseAnd() (ast.Expr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenAnd) {
		operator := p.previous()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = ast.Logical{Left: left, Operator: operator.Text, Right: right, Line: operator.Line}
	}
	return left, nil
}

func (p *Parser) parseEquality() (ast.Expr, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenBangEqual) || p.match(lexer.TokenEqualEqual) {
		operator := p.previous()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = ast.Binary{Left: left, Operator: operator.Text, Right: right, Line: operator.Line}
	}
	return left, nil
}

func (p *Parser) parseComparison() (ast.Expr, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.check(lexer.TokenGreater) || p.check(lexer.TokenGreaterEqual) ||
		p.check(lexer.TokenLess) || p.check(lexer.TokenLessEqual) {
		operator := p.advance()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = ast.Binary{Left: left, Operator: operator.Text, Right: right, Line: operator.Line}
	}
	return left, nil
}

func (p *Parser) parseTerm() (ast.Expr, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenMinus) || p.match(lexer.TokenPlus) {
		operator := p.previous()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = ast.Binary{Left: left, Operator: operator.Text, Right: right, Line: operator.Line}
	}
	return left, nil
}

func (p *Parser) parseFactor() (ast.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenSlash) || p.match(lexer.TokenStar) || p.match(lexer.TokenPercent) {
		operator := p.previous()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = ast.Binary{Left: left, Operator: operator.Text, Right: right, Line: operator.Line}
	}
	return left, nil
}

func (p *Parser) parseUnary() (ast.Expr, error) {
	if p.match(lexer.TokenBang) || p.match(lexer.TokenMinus) {
		operator := p.previous()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return ast.Unary{Operator: operator.Text, Right: right, Line: operator.Line}, nil
	}
	return p.parseCall()
}

func (p *Parser) parseCall() (ast.Expr, error) {
	expression, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		if p.match(lexer.TokenLeftParen) {
			expression, err = p.finishCall(expression)
			if err != nil {
				return nil, err
			}
		} else if p.match(lexer.TokenLeftBracket) {
			indexToken := p.previous()
			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if _, err := p.consume(lexer.TokenRightBracket, "expected ']' after index"); err != nil {
				return nil, err
			}
			expression = ast.Index{Object: expression, Index: index, Line: indexToken.Line}
		} else if p.match(lexer.TokenDot) {
			name, err := p.consume(lexer.TokenIdentifier, "expected property name after '.'")
			if err != nil {
				return nil, err
			}
			expression = ast.Get{Object: expression, Name: name.Text, Line: name.Line}
		} else {
			break
		}
	}
	return expression, nil
}

func (p *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	token := p.previous()
	arguments := make([]ast.Expr, 0)
	if !p.check(lexer.TokenRightParen) {
		for {
			argument, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, argument)
			if !p.match(lexer.TokenComma) {
				break
			}
		}
	}
	if _, err := p.consume(lexer.TokenRightParen, "expected ')' after arguments"); err != nil {
		return nil, err
	}
	return ast.Call{Callee: callee, Arguments: arguments, Line: token.Line}, nil
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	if p.match(lexer.TokenTrue) {
		token := p.previous()
		return ast.Boolean{Value: true, Line: token.Line}, nil
	}
	if p.match(lexer.TokenFalse) {
		token := p.previous()
		return ast.Boolean{Value: false, Line: token.Line}, nil
	}
	if p.match(lexer.TokenNil) {
		token := p.previous()
		return ast.Nil{Line: token.Line}, nil
	}
	if p.match(lexer.TokenFun) {
		return p.parseLambda()
	}
	if p.match(lexer.TokenInteger) {
		token := p.previous()
		value, _ := strconv.ParseInt(token.Value, 10, 64)
		return ast.Integer{Value: value, Line: token.Line}, nil
	}
	if p.match(lexer.TokenFloat) {
		token := p.previous()
		value, _ := strconv.ParseFloat(token.Value, 64)
		return ast.Float{Value: value, Line: token.Line}, nil
	}
	if p.match(lexer.TokenString) {
		token := p.previous()
		return ast.String{Value: token.Value, Line: token.Line}, nil
	}
	if p.match(lexer.TokenIdentifier) {
		token := p.previous()
		return ast.Variable{Name: token.Text, Line: token.Line}, nil
	}
	if p.match(lexer.TokenLeftParen) {
		token := p.previous()
		expression, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(lexer.TokenRightParen, "expected ')' after expression"); err != nil {
			return nil, err
		}
		_ = token
		return expression, nil
	}
	if p.match(lexer.TokenLeftBracket) {
		token := p.previous()
		items := make([]ast.Expr, 0)
		if !p.check(lexer.TokenRightBracket) {
			for {
				item, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				items = append(items, item)
				if !p.match(lexer.TokenComma) {
					break
				}
			}
		}
		if _, err := p.consume(lexer.TokenRightBracket, "expected ']' after list"); err != nil {
			return nil, err
		}
		return ast.List{Items: items, Line: token.Line}, nil
	}
	if p.match(lexer.TokenLeftBrace) {
		token := p.previous()
		entries := make([]ast.MapEntry, 0)
		if !p.check(lexer.TokenRightBrace) {
			for {
				key, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				if _, err := p.consume(lexer.TokenColon, "expected ':' after map key"); err != nil {
					return nil, err
				}
				value, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				entries = append(entries, ast.MapEntry{Key: key, Value: value})
				if !p.match(lexer.TokenComma) {
					break
				}
			}
		}
		if _, err := p.consume(lexer.TokenRightBrace, "expected '}' after map"); err != nil {
			return nil, err
		}
		return ast.Map{Entries: entries, Line: token.Line}, nil
	}
	token := p.current()
	if token.Kind == lexer.TokenEnd {
		return nil, fmt.Errorf("expected expression at end of input")
	}
	return nil, fmt.Errorf("unexpected token %q at line %d", token.Text, token.Line)
}

func (p *Parser) parseLambda() (ast.Expr, error) {
	token := p.previous()
	if _, err := p.consume(lexer.TokenLeftParen, "expected '(' after 'fun' in anonymous function"); err != nil {
		return nil, err
	}
	parameters := make([]string, 0)
	if !p.check(lexer.TokenRightParen) {
		for {
			parameter, err := p.consume(lexer.TokenIdentifier, "expected parameter name")
			if err != nil {
				return nil, err
			}
			parameters = append(parameters, parameter.Text)
			if !p.match(lexer.TokenComma) {
				break
			}
		}
	}
	if _, err := p.consume(lexer.TokenRightParen, "expected ')' after parameters"); err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return ast.Lambda{Parameters: parameters, Body: body, Line: token.Line}, nil
}

func (p *Parser) consumeOptionalSeparator() {
	p.match(lexer.TokenSemicolon)
	p.match(lexer.TokenNewline)
}

func (p *Parser) checkSeparator() bool {
	return p.check(lexer.TokenSemicolon) || p.check(lexer.TokenNewline)
}

func (p *Parser) skipSeparators() {
	for p.match(lexer.TokenSemicolon) || p.match(lexer.TokenNewline) {
	}
}

func (p *Parser) match(kind lexer.TokenKind) bool {
	if p.check(kind) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) consume(kind lexer.TokenKind, message string) (lexer.Token, error) {
	if p.check(kind) {
		return p.advance(), nil
	}
	token := p.current()
	return lexer.Token{}, fmt.Errorf("%s at line %d", message, token.Line)
}

func (p *Parser) check(kind lexer.TokenKind) bool {
	return p.current().Kind == kind
}

func (p *Parser) advance() lexer.Token {
	if p.position < len(p.tokens) {
		p.position++
	}
	return p.tokens[p.position-1]
}

func (p *Parser) previous() lexer.Token {
	return p.tokens[p.position-1]
}

func (p *Parser) current() lexer.Token {
	return p.tokens[p.position]
}

func (p *Parser) line() int {
	if p.position > 0 {
		return p.previous().Line
	}
	return 1
}

func (p *Parser) mapLiteralAhead() bool {
	if !p.check(lexer.TokenLeftBrace) || p.position+1 >= len(p.tokens) {
		return false
	}
	next := p.tokens[p.position+1].Kind
	return next == lexer.TokenString || next == lexer.TokenRightBrace
}
