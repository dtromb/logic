package logic

import (
	"fmt"
	"io"
	"bufio"
	"testing"
)

type BasicReader struct{
	data []byte
	pos int
}

type ConsoleWriter struct{}

func (br *BasicReader) Read(p []byte) (n int, err error) {
	l := len(p)
	if l > len(br.data)-br.pos {
		l = len(br.data)-br.pos
	}
	if l == 0 {
		return 0, io.EOF
	}
	copy(p, br.data[br.pos:br.pos+l])
	br.pos += l
	return l, nil
}

func StringReader(src string) io.Reader {
	return &BasicReader{data: []byte(src)}	
}

func (cw *ConsoleWriter) Write(p []byte) (n int, err error) {
	return fmt.Print(string(p))
}

func TestIO(t *testing.T) {
	source := CreateBasicParticleSource()
	var varX NamedParticle
	varX = source.GetVariableNamed("x")
	varY := source.GetVariableNamed("y")
	var qEx Name
	qEx = source.GetQuantifier("A")
	predEq := source.GetPredicateName("=")
	opImpl := source.GetOperator("->")
	predFoo := source.GetPredicateName("Foo")
	pred := source.Get(QUANTIFIED_PREDICATE, qEx, varX, 
				source.GetPredicateExpression(opImpl, 
					source.GetAtomicPredicate(predFoo,varX),
					source.GetAtomicPredicate(predEq,varX,varY)))
	out := &ConsoleWriter{}
	wout := GetStandardWriter()
	wout.Write(pred, out)
	fmt.Println()
	expected := "A$x:{->:Foo[$x],=[$x,$y]}"
	var rin LogicReader
	rin = GetStandardReader()
	in := bufio.NewReader(StringReader(expected))
	p, err := rin.ReadPredicate(source,in)
	if err != nil {
		t.Error(err)
	} else {
		wout.Write(p, out)
		fmt.Println()
	}
}