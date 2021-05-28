package runtime

import (
	"fmt"
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/runtime/operator"
	"mokapi/providers/workflow/token"
)

type treeNode struct {
	x  *treeNode
	y  *treeNode
	op token.Token
	o  interface{}
}

type binaryVisitor struct {
	outer  *visitor
	binary *ast.Binary
	tree   *treeNode
	ops    []token.Token
	n      int
}

func newBinaryVisitor(binary *ast.Binary, outer *visitor) *binaryVisitor {
	b := &binaryVisitor{outer: outer, binary: binary, n: 1}
	b.ops = append(b.ops, binary.Op)
	return b
}

func (v *binaryVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		switch n := e.(type) {
		case *ast.Binary:
			v.ops = append(v.ops, n.Op)
			v.n++
			return v
		}
		return v.outer.Visit(e)
	}

	if v.binary.Rhs == nil {
		return nil
	}

	v.n--

	i := len(v.ops) - 1
	op := v.ops[i]
	v.ops = v.ops[:i]

	if v.tree == nil {
		y := &treeNode{o: v.outer.stack.Pop()}
		x := &treeNode{o: v.outer.stack.Pop()}
		v.tree = &treeNode{x: x, y: y, op: op}
	} else {
		if v.tree.op.Precedence() < op.Precedence() {
			x := &treeNode{o: v.outer.stack.Pop()}
			n := &treeNode{x: x, y: v.tree.x, op: op}
			v.tree.x = n
		} else {
			x := &treeNode{o: v.outer.stack.Pop()}
			n := &treeNode{x: x, y: v.tree, op: op}
			v.tree = n
		}
	}

	if v.n == 0 {
		// finished building expression tree
		val, err := v.tree.eval()
		if err != nil {
			// TODO
			//v.outer.AddError(v.binary.Pos(), err.Error())
		} else {
			v.outer.stack.Push(val)
		}
	}

	return nil
}

func (n treeNode) eval() (interface{}, error) {
	if n.op == token.ILLEGAL {
		return n.o, nil
	}
	x, err := n.x.eval()
	if err != nil {
		return nil, err
	}
	y, err := n.y.eval()
	if err != nil {
		return nil, err
	}

	switch n.op {
	case token.ADD:
		return operator.Add(x, y)
	case token.SUB:
		return operator.Substract(x, y)
	case token.MUL:
		return operator.Multiply(x, y)
	case token.QUO:
		return operator.Divide(x, y)
	case token.REM:
		return operator.Modulo(x, y)
	case token.EQL:
		r, err := operator.Compare(x, y)
		if err != nil {
			return 0, err
		}
		return r == 0, nil
	case token.NEQ:
		r, err := operator.Compare(x, y)
		if err != nil {
			return 0, err
		}
		return r != 0, nil
	case token.LSS:
		r, err := operator.Compare(x, y)
		if err != nil {
			return 0, err
		}
		return r == -1, nil
	case token.GTR:
		r, err := operator.Compare(x, y)
		if err != nil {
			return 0, err
		}
		return r == 1, nil
	case token.LEQ:
		r, err := operator.Compare(x, y)
		if err != nil {
			return 0, err
		}
		return r <= 0, nil
	case token.GEQ:
		r, err := operator.Compare(x, y)
		if err != nil {
			return 0, err
		}
		return r >= 0, nil
	case token.LAND:
		return operator.And(x, y)
	case token.LOR:
		return operator.Or(x, y)
	}

	return 0, fmt.Errorf("unsupported operator %q", n.op)
}
