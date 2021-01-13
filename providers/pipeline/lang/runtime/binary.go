package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
)

type treeNode struct {
	x  *treeNode
	y  *treeNode
	op token.Token
	o  types.Object
}

func (n treeNode) eval() (types.Object, error) {
	if n.op == 0 {
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
	return x.InvokeOp(n.op, y)
}

type binaryVisitor struct {
	stack  *stack
	outer  visitor
	binary *ast.Binary
	tree   *treeNode
	ops    []token.Token
	n      int
}

func newBinaryVisitor(binary *ast.Binary, stack *stack, outer visitor) *binaryVisitor {
	b := &binaryVisitor{stack: stack, outer: outer, binary: binary, n: 1}
	b.ops = append(b.ops, binary.Op)
	return b
}

func (v *binaryVisitor) Visit(node ast.Node) ast.Visitor {
	if v.outer.hasErrors() {
		return nil
	}
	if node != nil {
		switch n := node.(type) {
		case *ast.Binary:
			v.ops = append(v.ops, n.Op)
			v.n++
			return v
		}
		return v.outer.Visit(node)
	}

	if v.binary.Rhs == nil {
		return nil
	}

	v.n--

	i := len(v.ops) - 1
	op := v.ops[i]
	v.ops = v.ops[:i]

	if v.tree == nil {
		y := &treeNode{o: v.stack.Pop()}
		x := &treeNode{o: v.stack.Pop()}
		v.tree = &treeNode{x: x, y: y, op: op}
	} else {
		if v.tree.op.Precedence() < op.Precedence() {
			x := &treeNode{o: v.stack.Pop()}
			n := &treeNode{x: x, y: v.tree.x, op: op}
			v.tree.x = n
		} else {
			x := &treeNode{o: v.stack.Pop()}
			n := &treeNode{x: x, y: v.tree, op: op}
			v.tree = n
		}
	}

	if v.n == 0 {
		// finished building expression tree
		val, err := v.tree.eval()
		if err != nil {
			v.outer.addError(err)
		} else {
			v.stack.Push(val)
		}
	}

	return nil
}

func (v *binaryVisitor) buildTree(x types.Object, y types.Object) {

}
