package yggdrasil

import (
	"context"
	"errors"
)

var (
	ErrPathInvalid  = errors.New("Invalid Path")
	ErrPathNotFound = errors.New("Not Found Path")
)

type Path interface{}

func NewPath(value interface{}) Path {
	return Path(value)
}

type Closeable interface {
	Close()
}

type Reference interface{}

func NewReference(value interface{}) Reference {
	return Reference(value)
}

type Tree interface {
	Reference(Path) (Reference, error)
	Close()
}

type emptyTree struct{}

func (emptyTree) Reference(Path) (Reference, error) {
	return nil, ErrPathNotFound
}

func (emptyTree) Close() {}

type referenceTree struct {
	context.Context
	parent    Tree
	path      Path
	reference Reference
}

func (t *referenceTree) Reference(path Path) (Reference, error) {
	if err := t.Context.Err(); err != nil {
		return nil, err
	}
	if t.path == path {
		return t.reference, nil
	}
	return t.parent.Reference(path)
}

func (t *referenceTree) Close() {
	if closeable, is := t.reference.(Closeable); is {
		closeable.Close()
	}
	t.path = nil
	if t.parent != nil {
		t.parent.Close()
	}
}

type Root struct {
	path      Path
	reference Reference
}

func newRoot(path Path, reference Reference) Root {
	return Root{path: path, reference: reference}
}

type Roots struct {
	register []Root
}

func (r *Roots) Register(path Path, reference Reference) error {
	if path == nil {
		return ErrPathInvalid
	}
	r.register = append(r.register, newRoot(path, reference))
	return nil
}

func NewRoots() Roots {
	return Roots{}
}

func (r *Roots) newTree(ctx context.Context) Tree {
	var tree Tree = emptyTree{}
	for _, root := range r.register {
		tree = &referenceTree{
			Context:   ctx,
			parent:    tree,
			path:      root.path,
			reference: root.reference,
		}
	}
	return tree
}

func (r *Roots) NewTreeDefault() Tree {
	return r.newTree(context.Background())
}

func (r *Roots) NewTree(ctx context.Context) Tree {
	return r.newTree(ctx)
}
