package logic

import "github.com/dtromb/logic/hash"

type ParticleType int
const (
	NAME						ParticleType = iota
	VARIABLE
	VARIABLE_NAME
	FUNCTION_NAME
	PREDICATE_NAME
	OPERATOR
	FUNCTION_EXPRESSION			// Term: tuple(Term)
	ATOMIC_PREDICATE			// Pred: tuple(Term)
	PREDICATE_COMPREHENSION		// Term: tuple(Pred)
	PREDICATE_EXPRESSION		// Pred: tuple(Pred)	
	QUANTIFIER
	QUANTIFIED_TERM
	QUANTIFIED_PREDICATE
)
func (pt ParticleType) String() string {
	switch(pt) {
		case NAME: return "name"
		case VARIABLE: return "variable"
		case VARIABLE_NAME: return "variable-name"
		case FUNCTION_NAME: return "function-name"
		case PREDICATE_NAME: return "predicate-name"
		case OPERATOR: return "operator"
		case FUNCTION_EXPRESSION: return "function-expression"		
		case ATOMIC_PREDICATE: return "atomic-predicate"
		case PREDICATE_COMPREHENSION: return "predicate-comprehension"
		case PREDICATE_EXPRESSION: return "predicate-expression"
		case QUANTIFIER: return "quantifier"
		case QUANTIFIED_TERM: return "quantified-term"
		case QUANTIFIED_PREDICATE: return "quantified-predicate"
	}
	return "<unknown>"
}

type ParticleSource interface {
	Get(ptype ParticleType, parts ...Particle) Particle
	GetName(nameType ParticleType, name string) Name
	GetVariableName(name string) Name
	GetFunctionName(name string) Name
	GetPredicateName(name string) Name
	GetVariable(name Name) NamedParticle
	GetVariableNamed(name string) NamedParticle
	GetOperator(name string) Name
	GetQuantifier(name string) Name
	GetTuple(nameType ParticleType, head Name, args ...Particle) TupleParticle
	GetFunctionExpression(funcName Name, terms ...Particle) TupleParticle
	GetAtomicPredicate(predName Name, terms ...Particle) TupleParticle
	GetPredicateExpression(operator Name, args ...Particle) TupleParticle
	GetPredicateComprehension(operator Name, args ...Particle) TupleParticle
	GetQuantifiedTerm(quantifier Name, variable NamedParticle, arg Particle) QuantifiedParticle
	GetQuantifiedPredicate(quantifier Name, variable NamedParticle, arg Particle) QuantifiedParticle
}

type Particle interface {
	Type() ParticleType
	Length() int
	Part(idx int) Particle
	Parts() []Particle
	Source() ParticleSource
	Term() bool
	Predicate() bool
	Name() bool
	Hash() uint64
	Equals(p Particle) bool
}

type Name interface {
	Particle
	String() string
}

type NamedParticle interface {
	Particle
	NameParticle() Name
	String() string
}

type TupleParticle interface {
	Particle
	Head() Name
	Arity() int
	Argument(idx int) Particle
	Arguments() []Particle
}

type QuantifiedParticle interface {
	Particle
	Quantifier() Name
	Variable() NamedParticle
	Argument() Particle
}

func HashParticleArray(arr []Particle) uint64 {
	var h uint64
	for _, p := range arr {
		h = (hash.FNV_PRIME * h) ^ p.Hash()
	}
	return h
}
