package logic

import (
	"fmt"
	"bufio"
	"errors"
	"io"
	"unicode"
)

type LogicWriter interface {
	Write(p Particle, w io.Writer) error
}

type LogicReader interface {
	ReadParticle(source ParticleSource, ptype ParticleType, in *bufio.Reader) (Particle, error)
	ReadTerm(source ParticleSource, in *bufio.Reader) (Particle, error)
	ReadPredicate(source ParticleSource, in *bufio.Reader) (Particle,error)
	IdentifierStart(r rune) bool
	IdentifierPart(r rune) bool
}

type StandardWriter struct {
	LogicWriter
	Chain LogicWriter
}

type StandardReader struct {
	LogicReader
	Chain LogicReader
	line int
	col int
	pos int
	errMsg string
	cTerm Particle
	cString string
	cIdentifier string
	cPred Particle
	cSource ParticleSource
}

func GetStandardWriter() *StandardWriter {
	sw := StandardWriter{}
	sw.Chain = &sw
	return &sw
}

func GetStandardReader() *StandardReader{
	sr := StandardReader{}
	sr.Chain = &sr
	return &sr
}

func (lw *StandardWriter) Write(p Particle, out io.Writer) error {
	switch(p.Type()) {
		case VARIABLE: {
			out.Write([]byte(fmt.Sprintf("$%s", p.(NamedParticle).String())))
		}
		case VARIABLE_NAME: {
			out.Write([]byte(fmt.Sprintf("'var:%s'", p.(Name).String())))
		}
		case FUNCTION_NAME: {
			out.Write([]byte(fmt.Sprintf("'func:%s'", p.(Name).String())))
		}
		case PREDICATE_NAME: {
			out.Write([]byte(fmt.Sprintf("'pred:%s'", p.(Name).String())))
		}
		case OPERATOR: {
			out.Write([]byte(fmt.Sprintf("'op:%s'", p.(Name).String())))
		}
		case QUANTIFIER: {
			out.Write([]byte(fmt.Sprintf("'quant:%s'", p.(Name).String())))
		}
		case FUNCTION_EXPRESSION: {
			tp := p.(TupleParticle)
			out.Write([]byte(fmt.Sprintf("%s(",tp.Head().String())))
			for i, t := range tp.Arguments() {
				lw.Chain.Write(t, out)
				if i < tp.Arity()-1 {
					out.Write([]byte(","))
				}
			}
			out.Write([]byte(")"))
		}	
		case ATOMIC_PREDICATE: {
			tp := p.(TupleParticle)
			out.Write([]byte(fmt.Sprintf("%s[",tp.Head().String())))
			for i, t := range tp.Arguments() {
				lw.Chain.Write(t, out)
				if i < tp.Arity()-1 {
					out.Write([]byte(","))
				}
			}
			out.Write([]byte("]"))
		}				
		case PREDICATE_COMPREHENSION: {
			tp := p.(TupleParticle)
			out.Write([]byte(fmt.Sprintf("{%s;",tp.Head().String())))
			for i, t := range tp.Arguments() {
				lw.Chain.Write(t, out)
				if i < tp.Arity()-1 {
					out.Write([]byte(","))
				}
			}
			out.Write([]byte("}"))
		}				
		case PREDICATE_EXPRESSION: {
			tp := p.(TupleParticle)
			out.Write([]byte(fmt.Sprintf("{%s:",tp.Head().String())))
			for i, t := range tp.Arguments() {
				lw.Chain.Write(t, out)
				if i < tp.Arity()-1 {
					out.Write([]byte(","))
				}
			}
			out.Write([]byte("}"))
		}
		case QUANTIFIED_TERM: fallthrough
		case QUANTIFIED_PREDICATE: {
			qp := p.(QuantifiedParticle)
			out.Write([]byte(qp.Quantifier().String()))
			lw.Chain.Write(qp.Variable(),out)
			out.Write([]byte(":"))
			lw.Chain.Write(qp.Argument(),out)
			out.Write([]byte(":"))
		}
		default: {
			return errors.New("attempt to write invalid particle type")
		}
	}
	return nil
}
	
func NamePrefix(nameType ParticleType) string {
	switch(nameType) {
		case VARIABLE_NAME: return "var"
		case QUANTIFIER: return "quant"
		case PREDICATE_NAME: return "pred"
		case FUNCTION_NAME: return "func"
	}
	return ""
}

func PredicateTupleMark(partType ParticleType) string {
	switch(partType) {
		case PREDICATE_COMPREHENSION: return ";"
		case PREDICATE_EXPRESSION: return ":"
	}
	return ""
}

func (sr *StandardReader) Error(err error) {
	sr.errMsg = err.Error()
	panic(sr.errMsg)
}

func (sr *StandardReader) NextString(in *bufio.Reader, expect string) string {
	if sr.cString != "" {
		if sr.cString != expect {
			sr.Error(errors.New(fmt.Sprintf("expected '%s'", expect)))
		}
		sr.cString = ""
		return expect
	}
	buf := make([]byte, len(expect))
	k, err := in.Read(buf)
	sr.col += k
	sr.pos += k
	if err != nil || k != len(expect) {
		if err == nil {
			err = errors.New("unexpected short read")
		}
		sr.Error(errors.New(fmt.Sprintf("expected '%s' (%s)", expect, err.Error())))
	}
	if string(buf) != expect {
		sr.Error(errors.New(fmt.Sprintf("expected '%s'", expect)))
	}
	return expect
}

func (sr *StandardReader) NextWS(in *bufio.Reader) int {
	var k int
	for {
		c, _, err := in.ReadRune()
		if err != nil {
			return k
		}
		if !unicode.IsSpace(c) {
			in.UnreadRune()
			break
		}
		sr.pos += 1
		if c == '\n' {
			sr.col = 1
			sr.line += 1
		} else {
			sr.col += 1
		}
		k += 1
	}
	return k
}

func (sr *StandardReader) TestPeek(in *bufio.Reader, c rune) bool {
	c0, _, err := in.ReadRune()
	if err != nil {
		return false
	}
	in.UnreadRune()
	return c == c0
}

func (sr *StandardReader) NextIdentifier(in *bufio.Reader) string {
	var id []rune	
	if sr.cIdentifier != "" {
		rv := sr.cIdentifier
		sr.cIdentifier = ""
		return rv
	}
	for {
		c, _, err := in.ReadRune()
		if err != nil {
			if len(id) != 0 {
				break
			}
			sr.Error(errors.New(fmt.Sprintf("expected identifier (%s)", err.Error())))
		}
		var ok bool
		if len(id) == 0 {
			ok = sr.IdentifierStart(c)
		} else {
			ok = sr.IdentifierPart(c)
		}
		if !ok {
			in.UnreadRune()
			if len(id) == 0 {
				break
			}
			sr.Error(errors.New(fmt.Sprintf("expected identifier (%s)", err.Error())))
		}
		sr.pos += 1
		sr.col += 1
		id = append(id, c)
	}
	return string(id)
}

func (sr *StandardReader) NextTerm(in *bufio.Reader) Particle {
	if sr.cTerm != nil {
		t := sr.cTerm
		sr.cTerm = nil
		return t
	}
	t, err := sr.Chain.ReadTerm(sr.cSource, in)
	if err != nil {
		sr.Error(errors.New(fmt.Sprintf("expected term (%s)",err.Error())))
	}
	return t
}

func (sr *StandardReader) NextPredicate(in *bufio.Reader) Particle {
	if sr.cPred != nil {
		t := sr.cPred
		sr.cPred= nil
		return t
	}
	t, err := sr.Chain.ReadPredicate(sr.cSource, in)
	if err != nil {
		sr.Error(errors.New(fmt.Sprintf("expected predicate (%s)",err.Error())))
	}
	return t
}

func (sr *StandardReader) ReadDelimitedTermList(in *bufio.Reader, ldelim rune, rdelim rune) []Particle {
	var args []Particle	
	sr.NextString(in, string(ldelim))
	sr.NextWS(in)
	if !sr.TestPeek(in, rdelim) {
		for {
			term := sr.NextTerm(in)
			args = append(args, term)
			sr.NextWS(in)
			if sr.TestPeek(in, rdelim) {
				break
			}
			sr.NextString(in, string(','))
			sr.NextWS(in)
		}
	}				
	return args				
}

func (sr *StandardReader) ReadParticle(source ParticleSource, ptype ParticleType, in *bufio.Reader) (rp Particle, re error) {
	defer func() {
		if r := recover(); r != nil {
			rp = nil
			re = errors.New(fmt.Sprintf("%s at %d:%d (pos=%d)", sr.errMsg, sr.line, sr.col, sr.pos))
		}
	}()
	switch(ptype) {
		case VARIABLE: {
			sr.NextString(in,"$")
			id := sr.NextIdentifier(in)
			return source.GetVariableNamed(id), nil
		}	
		case VARIABLE_NAME: fallthrough
		case FUNCTION_NAME: fallthrough
		case PREDICATE_NAME: fallthrough
		case OPERATOR: {
			sr.NextString(in, "'")
			key := sr.NextIdentifier(in)
		    if key != NamePrefix(ptype) {
				panic(fmt.Sprintf("expected '%s'", NamePrefix(ptype)))
			}	
			sr.NextString(in,":")
			id := sr.NextIdentifier(in)
			sr.NextString(in,"'")
			return source.GetName(ptype, id), nil
		}
		case FUNCTION_EXPRESSION: {
			name := sr.NextIdentifier(in)
			sr.NextWS(in)
			args := sr.ReadDelimitedTermList(in, '(', ')')
			return source.GetFunctionExpression(source.GetFunctionName(name), args...), nil
		}
		case ATOMIC_PREDICATE: {
			name := sr.NextIdentifier(in)
			sr.NextWS(in)
			args := sr.ReadDelimitedTermList(in, '[', ']')
			return source.GetAtomicPredicate(source.GetPredicateName(name), args...), nil
		}
		case PREDICATE_COMPREHENSION: fallthrough
		case PREDICATE_EXPRESSION: {
			var args []Particle
			sr.NextString(in,"{")
			sr.NextWS(in)
			op := sr.NextIdentifier(in)
			sr.NextWS(in)
			sr.NextString(in,PredicateTupleMark(ptype))
			sr.NextWS(in)
			if !sr.TestPeek(in, '}') {
				for {
					arg := sr.NextPredicate(in)
					args = append(args,arg)
					sr.NextWS(in)
					if sr.TestPeek(in,'}') {
						break
					}
					if !sr.TestPeek(in,',') {
						panic("expected ','")
					}
					sr.NextWS(in)
				}
			}
			return source.GetTuple(ptype, source.GetOperator(op), args...), nil
		}
		case QUANTIFIED_PREDICATE: fallthrough
		case QUANTIFIED_TERM: {
			q := sr.NextIdentifier(in)
			sr.NextWS(in)
			sr.NextString(in,"$")
			id := sr.NextIdentifier(in)
			sr.NextWS(in)
			sr.NextString(in,":")
			sr.NextWS(in)
			arg := sr.NextPredicate(in)
			sr.NextWS(in)
			sr.NextString(in,":")
			if ptype == QUANTIFIED_PREDICATE {
				return source.GetQuantifiedPredicate(source.GetQuantifier(q), 
				                                     source.GetVariableNamed(id),
													 arg), nil
			} else {
				return source.GetQuantifiedTerm(source.GetQuantifier(q), 
				                                source.GetVariableNamed(id),
											    arg), nil
			}
		}
		panic("unknown particle type")
	}
	return nil,nil
}

func (sr *StandardReader) ReadPredicate(source ParticleSource, in *bufio.Reader) (rp Particle, re error) {
	if sr.TestPeek(in, '{') {
		fmt.Println("PE")
		return sr.ReadParticle(source, PREDICATE_EXPRESSION, in)
	}
	fmt.Println("AP")
	return sr.ReadParticle(source, ATOMIC_PREDICATE, in)
}