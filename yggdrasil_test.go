package yggdrasil

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type testBase struct {
	name string
}

func newTestBase(name string) testBase {
	return testBase{name: name}
}

type testNewPath struct {
	testBase
	path interface{}
}

func newTestNewPath(name string, path interface{}) testNewPath {
	return testNewPath{
		testBase: newTestBase(name),
		path:     path,
	}
}

func TestNewPath(test *testing.T) {
	scenarios := []testNewPath{
		newTestNewPath(
			"with string path value",
			"string/mock/path",
		),
		newTestNewPath(
			"with int path value",
			10,
		),
		newTestNewPath(
			"with float path value",
			99.13,
		),
		newTestNewPath(
			"with struct path value",
			struct{}{},
		),
		newTestNewPath(
			"with map path value",
			make(map[interface{}]interface{}),
		),
	}

	for index, scenario := range scenarios {
		test.Run(
			fmt.Sprintf("[%d]-%s", index, scenario.name),
			func(t *testing.T) {
				path := NewPath(scenario.path)
				pathType, expectedType := reflect.TypeOf(path), reflect.TypeOf(scenario.path)
				if expectedType != pathType {
					t.Errorf("err_assert: assert=pathType expected=%T got=%T", scenario.path, path)
				}
				pathValue, expectedValue := reflect.ValueOf(path), reflect.ValueOf(scenario.path)
				if expectedValue != pathValue {
					t.Errorf("err_assert: assert=referenceValue expected=%v got=%v", scenario.path, path)
				}
			},
		)
	}
}

type testNewReference struct {
	testBase
	reference interface{}
}

func newTestNewReference(name string, reference interface{}) testNewReference {
	return testNewReference{
		testBase:  newTestBase(name),
		reference: reference,
	}
}

func TestNewReference(test *testing.T) {
	scenarios := []testNewReference{
		newTestNewReference(
			"with map reference value",
			make(map[interface{}]interface{}),
		),
	}

	for index, scenario := range scenarios {
		test.Run(
			fmt.Sprintf("[%d]-%s", index, scenario.name),
			func(t *testing.T) {
				reference := NewReference(scenario.reference)
				referenceType, expectedType := reflect.TypeOf(reference), reflect.TypeOf(scenario.reference)
				if expectedType != referenceType {
					t.Errorf("err_assert: assert=referenceType expected=%T got=%T", scenario.reference, reference)
				}
				referenceValue, expectedValue := reflect.ValueOf(reference), reflect.ValueOf(scenario.reference)
				if expectedValue != referenceValue {
					t.Errorf("err_assert: assert=referenceValue expected=%v got=%v", scenario.reference, reference)
				}
			},
		)
	}
}

type rootsScenario struct {
	testBase
	path      Path
	reference Reference
	err       error
}

func newRootsScenario(name string, path Path, reference Reference, err error) rootsScenario {
	return rootsScenario{
		testBase: testBase{
			name: name,
		},
		path:      path,
		reference: reference,
		err:       err,
	}
}

func TestRoots(test *testing.T) {
	scenarios := []rootsScenario{
		newRootsScenario(
			"Creates new Roots and register a pointer reference",
			NewPath("mock/pointer"), NewReference(new(struct{})),
			nil,
		),
		newRootsScenario(
			"Creates new Roots and register a scalar reference",
			NewPath("mock/scalar/string"), NewReference("scalar string"),
			nil,
		),
		newRootsScenario(
			"Creates new Roots and register a nil reference",
			NewPath("mock/scalar/nil"), nil,
			nil,
		),
		newRootsScenario(
			"Creates new Roots and try to register a reference with nil path",
			nil, NewReference(new(struct{})),
			ErrPathInvalid,
		),
	}

	for index, scenario := range scenarios {
		test.Run(
			fmt.Sprintf("[%d]-%s", index, scenario.name),
			func(t *testing.T) {
				roots := NewRoots()
				err := roots.Register(scenario.path, scenario.reference)
				if scenario.err != err {
					t.Errorf("err_assert: assert=errregister expected=%v got=%v", scenario.err, err)
				}

				for _, root := range roots.register {
					referenceType, expectedType := reflect.TypeOf(root.reference), reflect.TypeOf(scenario.reference)
					if expectedType != referenceType {
						t.Errorf("err_assert: assert=referenceType expected=%T got=%T", scenario.reference, root.reference)
					}
					referenceValue, expectedValue := reflect.ValueOf(root.reference), reflect.ValueOf(scenario.reference)
					if expectedValue != referenceValue {
						t.Errorf("err_assert: assert=referenceValue expected=%v got=%v", scenario.reference, root.reference)
					}
				}
			},
		)
	}
}

type closeableRef struct {
	open bool
}

func (c *closeableRef) Close() {
	c.open = false
}

type treeScenario struct {
	testBase
	roots      Roots
	references map[Path]Reference
	ctxMaker   func() (context.Context, context.CancelFunc)
	paths      []Path
	errors     []error
}

func (scenario *treeScenario) setup(t *testing.T) {
	roots := NewRoots()
	for path, reference := range scenario.references {
		err := roots.Register(path, reference)
		if err != nil {
			t.Errorf("err_assert: assert=errregister expected=nil got=%v", err)
		}
	}
	scenario.roots = roots
}

func newTreeScenario(
	name string,
	references map[Path]Reference,
	ctxMaker func() (context.Context, context.CancelFunc),
	paths []Path,
	errors []error,
) treeScenario {
	return treeScenario{
		testBase: testBase{
			name: name,
		},
		references: references,
		ctxMaker:   ctxMaker,
		paths:      paths,
		errors:     errors,
	}
}

func TestTree(test *testing.T) {
	scenarios := []treeScenario{
		newTreeScenario(
			"Creates a new Tree with all references registered and find all they",
			map[Path]Reference{
				NewPath("mock"):              NewReference(new(struct{})),
				NewPath("mock/scalarint"):    NewReference(102299),
				NewPath("mock/scalarfloat"):  NewReference(102299.9999),
				NewPath("mock/scalarstring"): NewReference("mystring"),
				NewPath("mock/scalar/time"):  NewReference(time.Now().UTC()),
				NewPath("mock/scalar/nil"):   NewReference(nil),
				NewPath("mock/closeable1"):   NewReference(&closeableRef{open: true}),
			},
			nil,
			[]Path{
				NewPath("mock"),
				NewPath("mock/scalarint"),
				NewPath("mock/scalarfloat"),
				NewPath("mock/scalarstring"),
				NewPath("mock/scalar/time"),
				NewPath("mock/scalar/nil"),
				NewPath("mock/scalarint"),
				NewPath("mock/scalarstring"),
				NewPath("mock/scalarfloat"),
				NewPath("mock"),
				NewPath("mock/scalar/time"),
				NewPath("mock/scalar/nil"),
				NewPath("mock/closeable1"),
			},
			nil,
		),
		newTreeScenario(
			"Creates a new Tree with some References but try to find one with a not found Path",
			map[Path]Reference{
				NewPath("mock/1"):          NewReference(new(struct{})),
				NewPath("mock/2"):          NewReference(new(struct{})),
				NewPath("mock/3"):          NewReference(new(struct{})),
				NewPath("mock/closeable2"): NewReference(&closeableRef{open: true}),
			},
			nil,
			[]Path{
				NewPath("mock"),
			},
			[]error{
				ErrPathNotFound,
			},
		),
		newTreeScenario(
			"Creates a new Tree with some References but try to find all with not found Paths",
			map[Path]Reference{
				NewPath("mock/1"):          NewReference(new(struct{})),
				NewPath("mock/2"):          NewReference(new(struct{})),
				NewPath("mock/3"):          NewReference(new(struct{})),
				NewPath("mock/closeable3"): NewReference(&closeableRef{open: true}),
			},
			nil,
			[]Path{
				NewPath("mock/notfound1"),
				NewPath("mock/notfound2"),
				NewPath("mock/notfound3"),
			},
			[]error{
				ErrPathNotFound,
				ErrPathNotFound,
				ErrPathNotFound,
			},
		),
		newTreeScenario(
			"Creates a new Tree with no References and try to find a bunch os Paths",
			map[Path]Reference{},
			nil,
			[]Path{
				NewPath("mock/notfound1"),
				NewPath("mock/notfound2"),
				NewPath("mock/notfound3"),
			},
			[]error{
				ErrPathNotFound,
				ErrPathNotFound,
				ErrPathNotFound,
			},
		),
		newTreeScenario(
			"Creates a new Tree with some References and a Context and find all Paths",
			map[Path]Reference{
				NewPath("mock/1"):          NewReference(new(struct{})),
				NewPath("mock/2"):          NewReference(new(struct{})),
				NewPath("mock/closeable4"): NewReference(&closeableRef{open: true}),
			},
			func() (context.Context, context.CancelFunc) {
				return context.Background(), nil
			},
			[]Path{
				NewPath("mock/1"),
				NewPath("mock/2"),
				NewPath("mock/closeable4"),
			},
			nil,
		),
		newTreeScenario(
			"Creates a new Tree with some References and already cancelled Context",
			map[Path]Reference{
				NewPath("mock/1"):          NewReference(new(struct{})),
				NewPath("mock/2"):          NewReference(new(struct{})),
				NewPath("mock/closeable5"): NewReference(&closeableRef{open: true}),
			},
			func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			[]Path{
				NewPath("mock/1"),
				NewPath("mock/2"),
				NewPath("mock/closeable5"),
			},
			[]error{
				context.Canceled,
				context.Canceled,
				context.Canceled,
			},
		),
		newTreeScenario(
			"Creates a new Tree with some References and already dealined Context",
			map[Path]Reference{
				NewPath("mock/1"):          NewReference(new(struct{})),
				NewPath("mock/2"):          NewReference(new(struct{})),
				NewPath("mock/closeable6"): NewReference(&closeableRef{open: true}),
			},
			func() (context.Context, context.CancelFunc) {
				deadLineTime := time.Now().UTC()
				ctx, cancel := context.WithDeadline(context.Background(), deadLineTime)
				for deadLineTime.After(time.Now().UTC()) {
				}
				return ctx, cancel
			},
			[]Path{
				NewPath("mock/1"),
				NewPath("mock/2"),
				NewPath("mock/closeable6"),
			},
			[]error{
				context.DeadlineExceeded,
				context.DeadlineExceeded,
				context.DeadlineExceeded,
			},
		),
		newTreeScenario(
			"Creates a new Tree with some References and already timedout Context",
			map[Path]Reference{
				NewPath("mock/1"):          NewReference(new(struct{})),
				NewPath("mock/2"):          NewReference(new(struct{})),
				NewPath("mock/closeable7"): NewReference(&closeableRef{open: true}),
			},
			func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
				time.Sleep(time.Millisecond * 2)
				return ctx, cancel
			},
			[]Path{
				NewPath("mock/1"),
				NewPath("mock/2"),
				NewPath("mock/closeable7"),
			},
			[]error{
				context.DeadlineExceeded,
				context.DeadlineExceeded,
				context.DeadlineExceeded,
			},
		),
	}

	for index, scenario := range scenarios {
		test.Run(
			fmt.Sprintf("[%d]-%s", index, scenario.name),
			func(t *testing.T) {
				scenario.setup(t)

				var (
					tree   Tree
					ctx    context.Context
					cancel context.CancelFunc
				)
				if scenario.ctxMaker == nil {
					tree = scenario.roots.NewTreeDefault()
					ctx = context.Background()
				} else {
					ctx, cancel = scenario.ctxMaker()
					tree = scenario.roots.NewTree(ctx)
				}
				if cancel != nil {
					defer cancel()
				}

				switch treeInstance := tree.(type) {
				case emptyTree:
					if len(scenario.references) > 0 {
						t.Errorf("err_assert: assert=emptyTreeWithReferences expected=%+v got=empty", len(scenario.references))
					}
				case *referenceTree:
					if ctx != treeInstance.Context {
						t.Errorf("err_assert: assert=referenceTreeContext expected=%+v got=%+v", ctx, treeInstance.Context)
					}
				default:
					t.Errorf("err_assert: assert=treeInvalidInstance expected=[emptyTree, referenceTree] got=%T", tree)
				}

				for index, path := range scenario.paths {
					reference, err := tree.Reference(path)
					var expectedErr error
					if len(scenario.errors) > index {
						expectedErr = scenario.errors[index]
					}
					if expectedErr != err {
						t.Errorf("err_assert: assert=treeReference expected=%v got=%v", expectedErr, err)
					}
					if expectedErr == nil {
						expectedReference, exists := scenario.references[path]
						if !exists {
							t.Error("err_assert: assert=scenarioReferenceNotFound expected=found got=notfound")
						}
						referenceType, expectedType := reflect.TypeOf(reference), reflect.TypeOf(expectedReference)
						if expectedType != referenceType {
							t.Errorf("err_assert: assert=treeReferenceType expected=%T got=%T", expectedReference, reference)
						}
						referenceValue, expectedValue := reflect.ValueOf(reference), reflect.ValueOf(expectedReference)
						if expectedValue != referenceValue {
							t.Errorf("err_assert: assert=treeReferenceValue expected=%v got=%v", expectedReference, reference)
						}
					}
				}

				tree.Close()
				// any another call to close does nothing
				tree.Close()
				tree.Close()
				for index, path := range scenario.paths {
					var expectedErr error
					if len(scenario.errors) > index {
						expectedErr = scenario.errors[index]
					}
					reference, err := tree.Reference(path)
					t.Logf("log: assert=closedTreeReference reference=%+v error=%+v expectedError=%+v", reference, err, expectedErr)
					if reference != nil {
						t.Errorf("err_assert: assert=closedTreeNonNilReference expected=nil got=%+v", reference)
					}
					if err == nil {
						t.Error("err_assert: assert=closedTreeErrReference expected=non nil got=nil")
					}
					if expectedErr != nil && expectedErr != err {
						t.Errorf("err_assert: assert=invalidClosedTreeErrReference expected=%v got=%v", expectedErr, err)
					}
				}

				for path, reference := range scenario.references {
					if closeable, is := reference.(*closeableRef); is {
						if closeable.open {
							t.Errorf("err_assert: assert=invalidCloseableState path=%v expected=closed got=open", path)
						}
					}
				}

			},
		)
	}
}
