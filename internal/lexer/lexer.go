package lexer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenKind int

const (
	TokenInteger TokenKind = iota
	TokenFloat
	TokenString
	TokenIdentifier
	TokenLet
	TokenFun
	TokenIf
	TokenElse
	TokenWhile
	TokenFor
	TokenIn
	TokenReturn
	TokenBreak
	TokenContinue
	TokenPrint
	TokenTrue
	TokenFalse
	TokenNil
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenPercent
	TokenBang
	TokenEqual
	TokenEqualEqual
	TokenBangEqual
	TokenLess
	TokenLessEqual
	TokenGreater
	TokenGreaterEqual
	TokenAnd
	TokenOr
	TokenLeftParen
	TokenRightParen
	TokenLeftBrace
	TokenRightBrace
	TokenLeftBracket
	TokenRightBracket
	TokenComma
	TokenDot
	TokenColon
	TokenSemicolon
	TokenNewline
	TokenEnd
	TokenImport
	TokenStruct
	TokenTry
	TokenCatch
	TokenThrow
)

type Token struct {
	Kind  TokenKind
	Text  string
	Value string
	Line  int
	Byte  int
}

var keywords = map[string]TokenKind{
	"let":      TokenLet,
	"fun":      TokenFun,
	"if":       TokenIf,
	"else":     TokenElse,
	"while":    TokenWhile,
	"for":      TokenFor,
	"in":       TokenIn,
	"return":   TokenReturn,
	"break":    TokenBreak,
	"continue": TokenContinue,
	"print":    TokenPrint,
	"import":   TokenImport,
	"struct":   TokenStruct,
	"try":      TokenTry,
	"catch":    TokenCatch,
	"throw":    TokenThrow,
	"true":     TokenTrue,
	"false":    TokenFalse,
	"nil":      TokenNil,
}

func (kind TokenKind) String() string {
	names := []string{
		"integer", "float", "string", "identifier", "let", "fun", "if", "else",
		"while", "for", "in", "return", "break", "continue", "print", "true",
		"false", "nil", "+", "-", "*", "/", "%", "!", "=", "==", "!=", "<",
		"<=", ">", ">=", "&&", "||", "(", ")", "{", "}", "[", "]", ",", ".",
		":", ";", "newline", "end", "import", "struct", "try", "catch", "throw",
	}
	if int(kind) >= 0 && int(kind) < len(names) {
		return names[kind]
	}
	return "unknown"
}

func Lex(source string) ([]Token, error) {
	tokens := make([]Token, 0)
	line := 1

	for position := 0; position < len(source); {
		character, size := utf8.DecodeRuneInString(source[position:])
		if character == utf8.RuneError && size == 1 {
			return nil, fmt.Errorf("invalid UTF-8 at line %d, byte %d", line, position)
		}

		if unicode.IsSpace(character) {
			if character == '\n' {
				tokens = append(tokens, Token{Kind: TokenNewline, Text: "\n", Line: line, Byte: position})
				line++
			}
			position += size
			continue
		}

		if strings.HasPrefix(source[position:], "//") {
			for position < len(source) && source[position] != '\n' {
				position++
			}
			continue
		}
		if strings.HasPrefix(source[position:], "/*") {
			startLine := line
			position += 2
			closed := false
			for position < len(source) {
				if strings.HasPrefix(source[position:], "*/") {
					position += 2
					closed = true
					break
				}
				if source[position] == '\n' {
					line++
				}
				_, charSize := utf8.DecodeRuneInString(source[position:])
				position += charSize
			}
			if !closed {
				return nil, fmt.Errorf("unterminated block comment starting on line %d", startLine)
			}
			continue
		}

		start := position
		switch character {
		case '+':
			tokens = append(tokens, Token{Kind: TokenPlus, Text: "+", Line: line, Byte: start})
		case '-':
			tokens = append(tokens, Token{Kind: TokenMinus, Text: "-", Line: line, Byte: start})
		case '*':
			tokens = append(tokens, Token{Kind: TokenStar, Text: "*", Line: line, Byte: start})
		case '/':
			tokens = append(tokens, Token{Kind: TokenSlash, Text: "/", Line: line, Byte: start})
		case '%':
			tokens = append(tokens, Token{Kind: TokenPercent, Text: "%", Line: line, Byte: start})
		case '!':
			kind := TokenBang
			text := "!"
			if position+1 < len(source) && source[position+1] == '=' {
				kind = TokenBangEqual
				text = "!="
				position++
			}
			tokens = append(tokens, Token{Kind: kind, Text: text, Line: line, Byte: start})
		case '=':
			kind := TokenEqual
			text := "="
			if position+1 < len(source) && source[position+1] == '=' {
				kind = TokenEqualEqual
				text = "=="
				position++
			}
			tokens = append(tokens, Token{Kind: kind, Text: text, Line: line, Byte: start})
		case '<':
			kind := TokenLess
			text := "<"
			if position+1 < len(source) && source[position+1] == '=' {
				kind = TokenLessEqual
				text = "<="
				position++
			}
			tokens = append(tokens, Token{Kind: kind, Text: text, Line: line, Byte: start})
		case '>':
			kind := TokenGreater
			text := ">"
			if position+1 < len(source) && source[position+1] == '=' {
				kind = TokenGreaterEqual
				text = ">="
				position++
			}
			tokens = append(tokens, Token{Kind: kind, Text: text, Line: line, Byte: start})
		case '&':
			if position+1 >= len(source) || source[position+1] != '&' {
				return nil, fmt.Errorf("unexpected '&' at line %d, byte %d", line, start)
			}
			tokens = append(tokens, Token{Kind: TokenAnd, Text: "&&", Line: line, Byte: start})
			position++
		case '|':
			if position+1 >= len(source) || source[position+1] != '|' {
				return nil, fmt.Errorf("unexpected '|' at line %d, byte %d", line, start)
			}
			tokens = append(tokens, Token{Kind: TokenOr, Text: "||", Line: line, Byte: start})
			position++
		case '(':
			tokens = append(tokens, Token{Kind: TokenLeftParen, Text: "(", Line: line, Byte: start})
		case ')':
			tokens = append(tokens, Token{Kind: TokenRightParen, Text: ")", Line: line, Byte: start})
		case '{':
			tokens = append(tokens, Token{Kind: TokenLeftBrace, Text: "{", Line: line, Byte: start})
		case '}':
			tokens = append(tokens, Token{Kind: TokenRightBrace, Text: "}", Line: line, Byte: start})
		case '[':
			tokens = append(tokens, Token{Kind: TokenLeftBracket, Text: "[", Line: line, Byte: start})
		case ']':
			tokens = append(tokens, Token{Kind: TokenRightBracket, Text: "]", Line: line, Byte: start})
		case ',':
			tokens = append(tokens, Token{Kind: TokenComma, Text: ",", Line: line, Byte: start})
		case '.':
			tokens = append(tokens, Token{Kind: TokenDot, Text: ".", Line: line, Byte: start})
		case ':':
			tokens = append(tokens, Token{Kind: TokenColon, Text: ":", Line: line, Byte: start})
		case ';':
			tokens = append(tokens, Token{Kind: TokenSemicolon, Text: ";", Line: line, Byte: start})
		case '"', '\'':
			quote := character
			startLine := line
			position += size
			var builder strings.Builder
			closed := false
			for position < len(source) {
				current, currentSize := utf8.DecodeRuneInString(source[position:])
				if current == quote {
					closed = true
					position += currentSize
					break
				}
				if current == '\\' {
					position += currentSize
					if position >= len(source) {
						break
					}
					escaped, escapedSize := utf8.DecodeRuneInString(source[position:])
					switch escaped {
					case 'n':
						builder.WriteByte('\n')
					case 'r':
						builder.WriteByte('\r')
					case 't':
						builder.WriteByte('\t')
					case '\\':
						builder.WriteByte('\\')
					case '"':
						builder.WriteByte('"')
					case '\'':
						builder.WriteByte('\'')
					default:
						builder.WriteRune(escaped)
					}
					position += escapedSize
					continue
				}
				if current == '\n' {
					line++
				}
				builder.WriteRune(current)
				position += currentSize
			}
			if !closed {
				return nil, fmt.Errorf("unterminated string starting on line %d", startLine)
			}
			tokens = append(tokens, Token{Kind: TokenString, Text: builder.String(), Value: builder.String(), Line: startLine, Byte: start})
			continue
		default:
			if character >= '0' && character <= '9' {
				startLine := line
				valueStart := position
				for position < len(source) {
					next, nextSize := utf8.DecodeRuneInString(source[position:])
					if (next < '0' || next > '9') && next != '.' {
						break
					}
					position += nextSize
				}
				text := source[valueStart:position]
				if strings.Count(text, ".") > 1 {
					return nil, fmt.Errorf("invalid number %q at line %d", text, startLine)
				}
				if strings.Contains(text, ".") {
					if _, err := strconv.ParseFloat(text, 64); err != nil {
						return nil, fmt.Errorf("invalid float %q at line %d", text, startLine)
					}
					tokens = append(tokens, Token{Kind: TokenFloat, Text: text, Value: text, Line: startLine, Byte: start})
				} else {
					if _, err := strconv.ParseInt(text, 10, 64); err != nil {
						return nil, fmt.Errorf("integer overflow at line %d", startLine)
					}
					tokens = append(tokens, Token{Kind: TokenInteger, Text: text, Value: text, Line: startLine, Byte: start})
				}
				continue
			}
			if unicode.IsLetter(character) || character == '_' {
				startLine := line
				valueStart := position
				position += size
				for position < len(source) {
					next, nextSize := utf8.DecodeRuneInString(source[position:])
					if !unicode.IsLetter(next) && !unicode.IsDigit(next) && next != '_' {
						break
					}
					position += nextSize
				}
				text := source[valueStart:position]
				kind, keyword := keywords[text]
				if !keyword {
					kind = TokenIdentifier
				}
				tokens = append(tokens, Token{Kind: kind, Text: text, Value: text, Line: startLine, Byte: start})
				continue
			}
			return nil, fmt.Errorf("unexpected character %q at line %d, byte %d", character, line, start)
		}
		position += size
	}

	tokens = append(tokens, Token{Kind: TokenEnd, Text: "end", Line: line, Byte: len(source)})
	return collapseNewlines(tokens), nil
}

func collapseNewlines(tokens []Token) []Token {
	result := make([]Token, 0, len(tokens))
	for index, token := range tokens {
		if token.Kind == TokenNewline {
			if index+1 < len(tokens) {
				next := tokens[index+1].Kind
				if next == TokenRightParen || next == TokenRightBracket ||
					next == TokenRightBrace || next == TokenComma {
					continue
				}
			}
			if len(result) == 0 || result[len(result)-1].Kind == TokenNewline ||
				result[len(result)-1].Kind == TokenLeftBrace ||
				result[len(result)-1].Kind == TokenComma ||
				result[len(result)-1].Kind == TokenLeftParen ||
				result[len(result)-1].Kind == TokenLeftBracket ||
				result[len(result)-1].Kind == TokenColon {
				continue
			}
		}
		result = append(result, token)
	}
	if len(result) > 0 && result[len(result)-1].Kind == TokenNewline {
		result = result[:len(result)-1]
	}
	return result
}
