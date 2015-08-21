package logic

import "fmt"
import "github.com/dtromb/logic/hash"

type BasicParticleSource struct {}

type BasicName struct {
	name string
	ptype ParticleType
	source ParticleSource
}

type BasicVariable struct {
	name Name
	source ParticleSource
}

type BasicTuple struct {
	head Name
	args []Particle
	ptype ParticleType
	source ParticleSource
 	hashcode uint64
}

type BasicQuantified struct {
	quantifier Name
	variable NamedParticle
	arg Particle
	ptype ParticleType
	source ParticleSource
}

func CreateBasicParticleSource() ParticleSource {
	return &BasicParticleSource{}
}

func (bs *BasicParticleSource) GetVariableName(name string) Name {
	return &BasicName{name: name, ptype: VARIABLE_NAME, source: bs}
}

func (bs *BasicParticleSource) GetFunctionName(name string) Name {
	return &BasicName{name: name, ptype: FUNCTION_NAME, source: bs}
}

func (bs *BasicParticleSource) GetPredicateName(name string) Name {
	return &BasicName{name: name, ptype: PREDICATE_NAME, source: bs}
}

func (bs *BasicParticleSource) GetOperator(name string) Name {
	return &BasicName{name: name, ptype: OPERATOR, source: bs}
}

func (bs *BasicParticleSource) GetQuantifier(name string) Name {
	return &BasicName{name: name, ptype: QUANTIFIER, source: bs}
}

func (bs *BasicParticleSource) GetVariableNamed(name string) NamedParticle {
	return &BasicVariable{name: bs.GetVariableName(name), source: bs}
}

func (bs *BasicParticleSource) GetVariable(name Name) NamedParticle {
	if name.Type() != VARIABLE_NAME {
		panic("name is not a variable name")
	}
	return &BasicVariable{name: name, source: bs}
}

func (bs *BasicParticleSource) GetName(nameType ParticleType, name string) Name {
	switch(nameType) {
		case VARIABLE_NAME: return bs.GetVariableName(name)
		case FUNCTION_NAME: return bs.GetFunctionName(name)
		case PREDICATE_NAME: return bs.GetFunctionName(name)
		case OPERATOR: return bs.GetOperator(name)
		case QUANTIFIER: return bs.GetQuantifier(name)
	}
	panic("type is not a name type")
}

func (bs *BasicParticleSource) GetFunctionExpression(funcName Name, terms ...Particle) TupleParticle {
	if funcName.Type() != FUNCTION_NAME {
		panic("name is not a function name")
	}
	t := make([]Particle, len(terms))
	for i, term := range terms {
		if !term.Term() {
			panic(fmt.Sprintf("expression argument %d is not a term", (i+1)))
		}
		t[i] = term
	}
	return &BasicTuple{ptype: FUNCTION_EXPRESSION, head: funcName, args: t, source: bs}
}

func (bs *BasicParticleSource) GetAtomicPredicate(predName Name, terms ...Particle) TupleParticle {
	if predName.Type() != PREDICATE_NAME {
		panic("name is not a predicate name")
	}
	t := make([]Particle, len(terms))
	for i, term := range terms {
		if !term.Term() {
			panic(fmt.Sprintf("predicate argument %d is not a term", (i+1)))
		}
		t[i] = term
	}
	return &BasicTuple{ptype: ATOMIC_PREDICATE, head: predName, args: t, source: bs}
}

func (bs *BasicParticleSource) GetPredicateExpression(op Name, args ...Particle) TupleParticle {
	if op.Type() != OPERATOR {
		panic("op is not an operator")
	}
	t := make([]Particle, len(args))
	for i, pred := range args {
		if !pred.Predicate() {
			panic(fmt.Sprintf("expression argument %d is not a predicate", (i+1)))
		}
		t[i] = pred
	}
	return &BasicTuple{ptype: PREDICATE_EXPRESSION, head: op, args: t, source: bs}
}

func (bs *BasicParticleSource) GetPredicateComprehension(op Name, args ...Particle) TupleParticle {
	if op.Type() != OPERATOR {
		panic("op is not an operator")
	}
	t := make([]Particle, len(args))
	for i, pred := range args {
		if !pred.Predicate() {
			panic(fmt.Sprintf("comprehension argument %d is not a predicate", (i+1)))
		}
		t[i] = pred
	}
	return &BasicTuple{ptype: PREDICATE_COMPREHENSION, head: op, args: t, source: bs}
}

func (bs *BasicParticleSource) GetTuple(tupleType ParticleType, head Name, args ...Particle) TupleParticle {
	switch(tupleType) {
		case ATOMIC_PREDICATE: return bs.GetAtomicPredicate(head, args...)
		case FUNCTION_EXPRESSION: return bs.GetAtomicPredicate(head, args...)
		case PREDICATE_EXPRESSION: return bs.GetPredicateExpression(head, args...)
		case PREDICATE_COMPREHENSION: return bs.GetPredicateComprehension(head, args...)
	}
	panic("type is not a tuple type")
}

func (bs *BasicParticleSource) GetQuantifiedTerm(quantifier Name, variable NamedParticle, arg Particle) QuantifiedParticle {
	if quantifier.Type() != QUANTIFIER {
		panic("first argument is not a quantifier")
	}
	if variable.Type() != VARIABLE {
		panic("second argument is not a variable")
	}
	if !arg.Predicate() {
		panic("third argument is not a predicate")
	}
	return &BasicQuantified{quantifier: quantifier, variable: variable, arg: arg, ptype: QUANTIFIED_TERM, source: bs}
}

func (bs *BasicParticleSource) GetQuantifiedPredicate(quantifier Name, variable NamedParticle, arg Particle) QuantifiedParticle {
	if quantifier.Type() != QUANTIFIER {
		panic("first argument is not a quantifier")
	}
	if variable.Type() != VARIABLE {
		panic("second argument is not a variable")
	}
	if !arg.Predicate() {
		panic("third argument is not a predicate")
	}
	return &BasicQuantified{quantifier: quantifier, variable: variable, arg: arg, ptype: QUANTIFIED_PREDICATE, source: bs}
}

func (bs *BasicParticleSource) Get(ptype ParticleType, parts ...Particle) Particle {
	switch(ptype) {
		case VARIABLE: {
			if len(parts) != 1 {
				panic("VARIABLE construction requires exactly one argument")
			}
			if n, ok := parts[0].(Name); ok {
				return bs.GetVariable(n)
			}
			panic("VARIABLE argument must be a Name")
		}
		case FUNCTION_EXPRESSION: fallthrough
		case ATOMIC_PREDICATE: fallthrough		
		case PREDICATE_COMPREHENSION: fallthrough
		case PREDICATE_EXPRESSION: {
			if len(parts) <= 0 {
				panic("tuple construction requires at least one argument")
			}
			if n, ok := parts[0].(Name); ok {
				return bs.GetTuple(ptype, n, parts[1:]...)
			}
			panic("tuple construction first argument must be a Name")
		}
		case QUANTIFIED_TERM: {
			if len(parts) != 3 {
				panic("quantified construction requires exactly three arguments")
			}
			if n, ok := parts[0].(Name); ok {
				if v, ok := parts[1].(NamedParticle); ok {
					return bs.GetQuantifiedTerm(n, v, parts[2])
				}
			}
			panic(fmt.Sprintf("quantified construction first two args must be Name and NamedPredicate (got %s and %s)",
					parts[0].Type().String(), parts[1].Type().String()))
		}
		case QUANTIFIED_PREDICATE: {
			if len(parts) != 3 {
				panic("quantified construction requires exactly three arguments")
			}	
			if n, ok := parts[0].(Name); ok {
				if v, ok := parts[1].(NamedParticle); ok {
					return bs.GetQuantifiedPredicate(n, v, parts[2])
				}
			}
			panic(fmt.Sprintf("quantified construction first two args must be Name and NamedPredicate (got %s and %s)",
					parts[0].Type().String(), parts[1].Type().String()))
		}
	}
	panic("invalid or name-typed particle requested")
}

func (n *BasicName) Type() ParticleType { return n.ptype }
func (n *BasicName) Length() int { return 0 }
func (n *BasicName) Part(idx int) Particle { return nil }
func (n *BasicName) Parts() []Particle { return []Particle{} }
func (n *BasicName) Source() ParticleSource { return n.source }
func (n *BasicName) Term() bool { return false }
func (n *BasicName) Predicate() bool { return false }
func (n *BasicName) Name() bool { return true }
func (n *BasicName) Hash() uint64 { return hash.HashString(n.name) ^ uint64(n.ptype)}
func (n *BasicName) Equals(p Particle) bool { 
	if p.Type() != n.ptype {
		return false
	}
	return p.(Name).String() == n.name
}
func (n *BasicName) String() string { return n.name }

func (v *BasicVariable) Type() ParticleType { return VARIABLE }
func (v *BasicVariable) Length() int { return 1 }
func (v *BasicVariable) Part(idx int) Particle { 
	if idx != 0 {
		return nil
	}
	return v.name
}
func (v *BasicVariable) Parts() []Particle { return []Particle{v.name} }
func (v *BasicVariable) Source() ParticleSource { return v.source }
func (v *BasicVariable) Term() bool { return true }
func (v *BasicVariable) Predicate() bool { return false }
func (v *BasicVariable) Name() bool { return false }
func (v *BasicVariable) Hash() uint64 { return v.Hash() ^ uint64(VARIABLE)}
func (v *BasicVariable) Equals(p Particle) bool { 
	if p.Type() != VARIABLE {
		return false
	}
	return p.(Name).String() == v.name.String()
}
func (v *BasicVariable) String() string { return v.name.String() }
func (v *BasicVariable) NameParticle() Name { return v.name }

func (t *BasicTuple) Type() ParticleType { return t.ptype }
func (t *BasicTuple) Length() int { return 1+len(t.args) }
func (t *BasicTuple) Part(idx int) Particle { 
	if idx == 0 {
		return t.head
	}
	if idx > len(t.args) {
		return nil
	}
	return t.args[idx-1]
}
func (t *BasicTuple) Parts() []Particle { return append([]Particle{t.head},t.args...) }
func (t *BasicTuple) Source() ParticleSource { return t.source }
func (t *BasicTuple) Term() bool { 
	switch(t.ptype) {
		case FUNCTION_EXPRESSION: return true
		case PREDICATE_COMPREHENSION: return true
	}
	return false
}
func (t *BasicTuple) Predicate() bool {
	switch(t.ptype) {
		case ATOMIC_PREDICATE: return true
		case PREDICATE_EXPRESSION: return true 
	}
	return false
}
func (t *BasicTuple) Name() bool { return false }
func (t *BasicTuple) Hash() uint64 { 
	if t.hashcode == 0 { 
		t.hashcode = (t.head.Hash() * 11) ^ HashParticleArray(t.args) 
	}
	return t.hashcode
}
func (t *BasicTuple) Equals(p Particle) bool { 
	if p.Type() != t.ptype {
		return false
	}
	if t == p {
		return true
	}
	if t.Hash() != p.Hash() {
		return true
	}
	tp, ok := p.(TupleParticle)
	if !ok {
		return false
	}
	if !tp.Head().Equals(t.head) {
		return false
	}
	if tp.Arity() != len(t.args) {
		return false
	}
	for i, a := range tp.Arguments() {
		if !a.Equals(t.args[i]) {
			return false
		}
	}
	return true
}	
func (t *BasicTuple) Head() Name { return t.head }
func (t *BasicTuple) Arity() int { return len(t.args) }
func (t *BasicTuple) Argument(idx int) Particle { return t.args[idx] }
func (t *BasicTuple) Arguments() []Particle { return append([]Particle{},t.args...)} 

func (q *BasicQuantified) Type() ParticleType { return q.ptype }
func (q *BasicQuantified) Length() int { return 3 }
func (q *BasicQuantified) Part(idx int) Particle { 
	switch(idx) {
		case 0: return q.quantifier
		case 1: return q.variable
		case 2: return q.arg
	}
	return nil
}
func (q *BasicQuantified) Parts() []Particle { return []Particle{q.quantifier, q.variable, q.arg} }
func (q *BasicQuantified) Source() ParticleSource { return q.source }
func (q *BasicQuantified) Term() bool { 
	switch(q.ptype) {
		case QUANTIFIED_TERM: return true
	} 
	return false
}
func (q *BasicQuantified) Predicate() bool { 
	switch(q.ptype) {
		case QUANTIFIED_PREDICATE: return true
	} 
	return false
}
func (q *BasicQuantified) Name() bool { return false }
func (q *BasicQuantified) Hash() uint64 { 
	return (q.quantifier.Hash()*3) ^ (q.variable.Hash()*5) ^ q.arg.Hash()
}
func (q *BasicQuantified) Equals(p Particle) bool { 
	if p.Type() != q.ptype {
		return false
	}
	qp, ok := p.(QuantifiedParticle) 
	if !ok {
		return false
	}
	return qp.Quantifier().Equals(q.quantifier) &&
	       qp.Variable().Equals(q.variable) &&
		   qp.Argument().Equals(q.arg)
}	
func (q *BasicQuantified) Quantifier() Name { return q.quantifier }
func (q *BasicQuantified) Variable() NamedParticle { return q.variable }
func (q *BasicQuantified) Argument() Particle { return q.arg }