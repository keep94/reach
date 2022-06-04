// This program prints out how all the numbers from 0 up to 200 exclusive can
// be written using 1, 2, 3, 4, and 5 along with different mathematical
// operators.
package main

import (
	"fmt"

	"github.com/keep94/gocombinatorics"
)

const (
	kMax = 1000000000 // operands must be less than this number
)

const (
	kNumExpressions = 200
)

// postfixEntry represents an entry in a postfix expression.
type postfixEntry struct {
	value int64 // The value of the entry
	op    byte  // '+', '-', '*', '/', '^' 0 means the entry is a value
}

func (s postfixEntry) String() string {
	if s.op != 0 {
		return fmt.Sprintf("%c", s.op)
	}
	return fmt.Sprintf("%d", s.value)
}

// postfix represents a postfix expression.
type postfix []postfixEntry

// Clear clears the postfix expression.
func (p *postfix) Clear() {
	*p = (*p)[:0]
}

// AppendValue appends a value to this expression.
func (p *postfix) AppendValue(x int64) {
	*p = append(*p, postfixEntry{value: x})
}

// AppendOp appends an operation to this expression.
func (p *postfix) AppendOp(op byte) {
	*p = append(*p, postfixEntry{op: op})
}

// Copy returns a copy of this postfix expression.
func (p postfix) Copy() postfix {
	result := make(postfix, len(p))
	copy(result, p)
	return result
}

// Eval evaluates this postfix expression. scratch is used to perform the
// calculation. Caller can declare a single []int64 slice and pass that same
// slice to multiple Eval() calls. Eval returns the value of the postfix
// expression and true or if the postfix expression can't be evaluated, it
// returns 0 and false. Reasons that a postfix expression can't be evaluated
// include division by zero, divisions that result in a non integer value etc.
func (p postfix) Eval(scratch *[]int64) (result int64, ok bool) {
	*scratch = (*scratch)[:0]
	for _, entry := range p {
		if entry.op == 0 {
			*scratch = append(*scratch, entry.value)
		} else {
			length := len(*scratch)
			second := (*scratch)[length-1]
			first := (*scratch)[length-2]
			*scratch = (*scratch)[:length-2]
			var answer int64
			var valid bool
			if entry.op == '+' {
				answer, valid = add(first, second)
			} else if entry.op == '-' {
				answer, valid = sub(first, second)
			} else if entry.op == '*' {
				answer, valid = mul(first, second)
			} else if entry.op == '/' {
				answer, valid = div(first, second)
			} else if entry.op == '^' {
				answer, valid = pow(first, second)
			} else {
				panic("Unrecognized op code")
			}
			if !valid {
				return
			}
			*scratch = append(*scratch, answer)
		}
	}
	if len(*scratch) != 1 {
		panic("Malformed postfix expression")
	}
	return (*scratch)[0], true
}

func checkRange(first, second int64) bool {
	if first <= -kMax || first >= kMax {
		return false
	}
	if second <= -kMax || second >= kMax {
		return false
	}
	return true
}

func add(first, second int64) (result int64, ok bool) {
	if !checkRange(first, second) {
		return
	}
	return first + second, true
}

func sub(first, second int64) (result int64, ok bool) {
	if !checkRange(first, second) {
		return
	}
	return first - second, true
}

func mul(first, second int64) (result int64, ok bool) {
	if !checkRange(first, second) {
		return
	}
	return first * second, true
}

func div(first, second int64) (result int64, ok bool) {
	if !checkRange(first, second) {
		return
	}
	if second == 0 || first%second != 0 {
		return
	}
	return first / second, true
}

func pow(first, second int64) (result int64, ok bool) {
	if second < 0 || second > 30 {
		return
	}
	product := int64(1)
	var valid bool
	for i := int64(0); i < second; i++ {
		product, valid = mul(product, first)
		if !valid {
			return
		}
	}
	return product, true
}

// The generator type generates postfix expressions
type generator struct {
	opsStream    gocombinatorics.Stream
	positsStream gocombinatorics.Stream
	valuesStream gocombinatorics.Stream
	ops          []int
	posits       []int
	values       []int
	actualOps    []byte
	actualValues []int64
	done         bool
}

// newGenerator returns a new generator that yields all the possible postfix
// expressions containing the given values and operations. Both values and
// ops must have a length of at least 1.
func newGenerator(values []int, ops []byte) *generator {
	if len(values) == 0 || len(ops) == 0 {
		panic("Length of both values and ops must be non-zero")
	}
	actualValues := make([]int64, len(values))
	for i := range values {
		actualValues[i] = int64(values[i])
	}
	actualOps := make([]byte, len(ops))
	copy(actualOps, ops)
	result := &generator{
		opsStream:    gocombinatorics.Product(len(ops), len(values)-1),
		positsStream: gocombinatorics.OpsPosits(len(values) - 1),
		valuesStream: gocombinatorics.Permutations(len(values), len(values)),
		ops:          make([]int, len(values)-1),
		posits:       make([]int, len(values)-1),
		values:       make([]int, len(values)),
		actualOps:    actualOps,
		actualValues: actualValues}
	result.opsStream.Next(result.ops)
	result.positsStream.Next(result.posits)
	result.valuesStream.Next(result.values)
	return result
}

// PostfixSize returns the size of the slice that caller must pass to Next()
// each time. This is 2*n - 1 where n is the length of the values slice
// passed to newGenerator.
func (g *generator) PostfixSize() int {
	return 2*len(g.actualValues) - 1
}

// Next stores the next postfix expression in p and returns true. If there
// are no more postfix expressions, Next leaves p unchanged and returns false.
// Caller must ensure that len(p) == g.PostfixSize().
func (g *generator) Next(p postfix) bool {
	if len(p) < g.PostfixSize() {
		panic("postfix slice passed to Next is too small")
	}
	if g.done {
		return false
	}
	g.populate(p)
	g.increment()
	return true
}

func (g *generator) increment() {
	if g.valuesStream.Next(g.values) {
		return
	}
	g.valuesStream.Reset()
	g.valuesStream.Next(g.values)

	if g.positsStream.Next(g.posits) {
		return
	}
	g.positsStream.Reset()
	g.positsStream.Next(g.posits)

	if g.opsStream.Next(g.ops) {
		return
	}
	g.done = true
}

func (g *generator) populate(p postfix) {
	p.Clear()
	valueIdx := 0
	positIdx := 0
	for valueIdx < len(g.values) || positIdx < len(g.posits) {
		if positIdx == len(g.posits) || valueIdx <= g.posits[positIdx] {
			p.AppendValue(g.actualValues[g.values[valueIdx]])
			valueIdx++
		} else {
			p.AppendOp(g.actualOps[g.ops[positIdx]])
			positIdx++
		}
	}
}

func main() {

	// expressions[0] contains an expression that evaluates to 0.
	// expressions[1] contains an expression that evaluates to 1
	// expressions[2] contains an expression that evaluates to 2 etc.
	// If expressions[n] is nil then there is no known expression that
	// evaluates to n.
	expressions := make([]postfix, kNumExpressions)

	g := newGenerator([]int{1, 2, 3, 4, 5}, []byte{'+', '-', '*', '/', '^'})
	p := make(postfix, g.PostfixSize())

	// Evaluate each possible expression.
	// Store that expression in the expressions array
	var scratch []int64
	for g.Next(p) {
		result, ok := p.Eval(&scratch)
		if !ok || result < 0 || result >= int64(len(expressions)) {
			continue
		}
		if expressions[result] == nil {
			expressions[result] = p.Copy()
		}
	}

	// Print the expression that evaluates to 0, 1, 2, 3, 4, ...
	for i := range expressions {
		fmt.Println(i, expressions[i])
	}
}
