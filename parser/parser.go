package parser

import (
	"bytes"
	"errors"
	"strings"
	"io"
)

type OutputLine struct {
	Signal   string
	FuncName string
	Args     []string
	Result   string
	Cover    []uint64
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) Parse() (*OutputLine, error) {
	line := &OutputLine{}
	tok, lit := p.scanIgnoreWhitespace()
	if tok == EOF || lit == "+" {
		return line, ErrEOF
	}
	if strings.Contains(lit, "Cover:") {
		line.Result = lit
		for {
			tok, lit = p.scanIgnoreWhitespace()
			if tok == NEWLINE {
				return line, nil
			}
		}
	}
	// Handle signals.
	if tok == SIGNAL {
		// Parse the line unchanged.
		var buf bytes.Buffer
		buf.WriteString(lit)
		for {
			tok, lit = p.scan()
			if tok == NEWLINE {
				break
			}
			buf.WriteString(lit)
		}
		line.Signal = buf.String()
		return line, nil
	} else {
		line.FuncName = lit
	}

	tok, lit = p.scanIgnoreWhitespace()
	if tok == OPEN_PAREN {
		// Read all the args up to CLOSE_PAREN
		for {
			tok, lit = p.scanIgnoreWhitespace()
			if tok == OPEN_PAREN {
				//If we encounter an open paren
				openParenCtr := 1
				var buf bytes.Buffer
				buf.WriteString(line.Args[len(line.Args)-1])
				buf.WriteString(lit)
				for {
					tok, lit = p.scan()

					if tok == OPEN_PAREN {
						buf.WriteString(lit)
						openParenCtr += 1
					} else if tok == CLOSE_PAREN {
						buf.WriteString(lit)
						openParenCtr -= 1
						if openParenCtr == 0 {
							break
						}
					} else {
						buf.WriteString(lit)
					}
				}
				line.Args[len(line.Args)-1] = buf.String()
			} else if tok == CLOSE_PAREN {
				break
			} else if tok == MEMADDR || tok == POINTER {
				line.Args = append(line.Args, lit)
				// Parse any struct arguments as a single arg.
			} else if tok == THROWAWAY {
				continue
			} else if tok == OPEN_BRACE {
				var buf bytes.Buffer
				buf.WriteString(lit)
				open_brace_counter := 1
				for {
					tok, lit = p.scan()

					if tok == OPEN_BRACE {
						buf.WriteString(lit)
						open_brace_counter++
					} else if tok == CLOSE_BRACE {
						buf.WriteString(lit)
						open_brace_counter--
						if (open_brace_counter == 0) {
							line.Args = append(line.Args, buf.String())
							break
						}
					} else if tok == MEMADDR || tok == POINTER {
						buf.WriteString(lit)
					} else {
						buf.WriteString(lit)
					}
				}
			} else if tok == OPEN_SQ {
				var buf bytes.Buffer
				buf.WriteString(lit)
				open_sq_counter := 1
				for {
					tok, lit = p.scan()
					if tok == OPEN_SQ {
						buf.WriteString(lit)
						open_sq_counter++
					} else if tok == CLOSE_SQ {
						buf.WriteString(lit)
						open_sq_counter--
						if (open_sq_counter == 0) {
							line.Args = append(line.Args, buf.String())
							break
						}
					} else if tok == MEMADDR || tok == POINTER{
						buf.WriteString(lit)
					} else {
						buf.WriteString(lit)
					}
				}
			} else if tok == STRING {
				line.Args = append(line.Args, lit)
			} else if tok != SEP {
				line.Args = append(line.Args, lit)
			}
		}
	}  else {
		return nil, errors.New("Expected OPEN_PAREN")
	}

	tok, lit = p.scanIgnoreWhitespace()
	if tok != EQUALS {
		return nil, errors.New("Expected EQUALS")
	}

	// Read everything after '=' until newline as the result.
	var result bytes.Buffer

	tok, lit = p.scanIgnoreWhitespace()
	result.WriteString(lit)
	line.Result = result.String()

	for {
		tok, lit = p.scanIgnoreWhitespace()
		if tok == NEWLINE {
			break
		}
	}
	return line, nil
}

func (o *OutputLine) Unparse() string {
	var buf bytes.Buffer
	if o.Signal != "" {
		buf.WriteString(o.Signal)
	} else {
		buf.WriteString(o.FuncName)
		buf.WriteString("(")
		for idx, arg := range o.Args {
			buf.WriteString(arg)
			if idx < len(o.Args)-1 {
				buf.WriteString(", ")
			} else {
				buf.WriteString(")")
			}
		}
		if len(o.Args) == 0 {
			buf.WriteString(")")
		}
		buf.WriteString(" = ")
		buf.WriteString(o.Result)
	}
	return buf.String()
}
