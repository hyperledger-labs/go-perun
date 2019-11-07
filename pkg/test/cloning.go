// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/pkg/test"

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

// isCloneable checks if the given type possesses a method `Clone`.  Receiver
// and return value can be values or references, e.g., with a method `func (*T)
// Clone() T`, the type `T` is considered cloneable.
func isCloneable(t reflect.Type) bool {
	kind := t.Kind()
	if kind != reflect.Struct && kind != reflect.Ptr {
		return false
	}

	// t may be a **struct
	baseType := t
	ptrType := reflect.PtrTo(t)
	for baseType.Kind() == reflect.Ptr {
		ptrType = baseType
		baseType = ptrType.Elem()
	}

	// check for clone method
	method, ok := ptrType.MethodByName("Clone")
	if !ok {
		return false
	}

	methodType := method.Type
	if numIn := methodType.NumIn(); numIn != 1 {
		return false
	}
	if inputType := methodType.In(0); inputType != ptrType {
		panic("This state never occurred during development")
	}
	if numOut := methodType.NumOut(); numOut != 1 {
		return false
	}
	if outputType := methodType.Out(0); outputType != baseType && outputType != ptrType {
		return false
	}

	return true
}

// checkCloneImpl recursively checks if the provided values could be clones and
// it returns an error if they cannot be.
//
// checkCloneImpl requires that the values referenced by v and w are deeply
// equal. Specifically, `reflect.DeepEqual(x, y)` must have had returned true,
// where
// * v = reflect.Value(x),
// * w = reflect.Value(y).
func checkCloneImpl(v, w reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		if v.Pointer() == 0 && w.Pointer() == 0 {
			panic("BUG: checkCloneImpl got nil inputs")
		}
		if v.Pointer() == w.Pointer() {
			return errors.New("Both arguments reference the same structure")
		}
	}

	if v.Kind() == reflect.Ptr && v.Elem().Kind() != reflect.Struct {
		panic("BUG: expected reference to struct, got reference to reference")
	}

	// get struct type
	baseType := v.Type()
	ptrType := reflect.PtrTo(baseType)
	if baseType.Kind() == reflect.Ptr {
		v = v.Elem()
		w = w.Elem()
		ptrType = baseType
		baseType = ptrType.Elem()
	}

	t := baseType

	for i := 0; i < baseType.NumField(); i++ {
		f := baseType.Field(i)
		kind := f.Type.Kind()
		left := v.Field(i)
		right := w.Field(i)
		// disallow some untested kinds
		if kind == reflect.Chan ||
			kind == reflect.Func || // disallow because of captured references
			kind == reflect.Map ||
			kind == reflect.String ||
			kind == reflect.UnsafePointer {
			panic(fmt.Sprintf("Implementation not tested with %v", kind))
		}

		// check for field tags
		tag, hasTag := f.Tag.Lookup("cloneable")
		// find unknown and misplaced tags
		if hasTag {
			if tag == "shallow" {
				if kind != reflect.Interface &&
					kind != reflect.Ptr &&
					kind != reflect.Slice {
					return errors.Errorf(
						"Expected field %v.%s with tag '%s' to be a "+
							"pointer or a slice, got kind %v",
						t, f.Name, tag, kind)
				}
			} else if tag == "shallowElements" {
				if kind != reflect.Array && kind != reflect.Slice {
					return errors.Errorf(
						"Expected field %v.%s with tag '%s' to be an array or "+
							"a slice, got kind %v",
						t, f.Name, tag, kind)
				}
			} else {
				return errors.Errorf(
					`Unknown tag 'cloneable:"%s"' on field %v.%s`,
					tag, t, f.Name)
			}
		}

		// check actual field contents
		if kind == reflect.Interface {
			if left.IsZero() && right.IsZero() {
				continue
			}

			if left.IsZero() != right.IsZero() {
				// reflect.DeepEqual() should detect this case
				panic(fmt.Sprintf(
					"Expected both interfaces to be zero or both non-zero, got %v, %v",
					left.InterfaceData(), right.InterfaceData()))
			}

			left = left.Elem()
			right = right.Elem()
			kindL := left.Type().Kind()
			kindR := right.Type().Kind()

			if kindL != reflect.Ptr && kindL != reflect.Struct {
				panic(
					fmt.Sprintf("Expected left kind ptr or struct, got %v",
						left.Type().Kind()))
			}
			if kindR != reflect.Ptr && kindR != reflect.Struct {
				panic(
					fmt.Sprintf("Expected right kind ptr or struct, got %v",
						right.Type().Kind()))
			}

			kind = left.Type().Kind()
		}

		if kind == reflect.Ptr || kind == reflect.Slice {
			p := left.Pointer()
			q := right.Pointer()
			if p != q && hasTag && tag == "shallow" {
				return errors.Errorf(
					"Expected fields %v.%s with tag '%s' to have same pointees",
					t, f.Name, tag)
			}
			// the length check below is necessary because all slices created
			// empty seem to reference the same address in memory
			if p == q && p != 0 &&
				(!hasTag || tag != "shallow") &&
				(kind == reflect.Ptr || left.Len() > 0) {
				return errors.Errorf(
					"Expected fields %v.%s to have different pointees",
					t, f.Name)
			}
			if p != q && kind == reflect.Ptr && isCloneable(f.Type) {
				err := checkCloneImpl(left.Elem(), right.Elem())
				if err != nil {
					return errors.Errorf(
						"Error in cloneable field %v.%s: %v",
						t, f.Name, err)
				}
			}
		}

		if kind == reflect.Array || kind == reflect.Slice {
			n := left.Len()
			for j := 0; j < n; j++ {
				kindJ := left.Index(j).Kind()
				if kindJ == reflect.Ptr || kindJ == reflect.Slice {
					p := left.Index(j).Pointer()
					q := right.Index(j).Pointer()
					if p != q && hasTag && tag == "shallowElements" {
						return errors.Errorf(
							"Expected elements %v.%s[%d] in slices with tag "+
								"'%s' to have same pointees",
							t, f.Name, j, tag)
					}
					if p == q && p != 0 && (!hasTag || tag != "shallowElements") {
						return errors.Errorf(
							"Expected elements %v.%s[%d] to have different pointees",
							t, f.Name, j)
					}
				} else if kindJ == reflect.Struct && isCloneable(f.Type.Elem()) {
					err := checkCloneImpl(left.Index(j), right.Index(j))
					if err != nil {
						return errors.Errorf(
							"Error in cloneable element %v.%s[%d]: %v",
							t, f.Name, j, err)
					}
				}
			}
		} else if kind == reflect.Struct && isCloneable(f.Type) {
			err := checkCloneImpl(left, right)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// checkClone checks if the two provided values could be clones. If they are
// not, an error is returned.
//
// checkClone initially calls reflect.DeepEqual and then checkCloneImpl tests
// recursively if the provided values could be clones.
func checkClone(p, q interface{}) error {
	if !reflect.DeepEqual(p, q) {
		return errors.New("Proper clones must be deeply equal")
	}
	if p == nil || q == nil {
		return errors.Errorf("Input must not be nil, got %v, %v", p, q)
	}

	tP := reflect.TypeOf(p)
	tQ := reflect.TypeOf(q)
	if tP != tQ || !isCloneable(tP) {
		return errors.Errorf("Input must be cloneable, type %v is not", tP)
	}

	v := reflect.ValueOf(p)
	w := reflect.ValueOf(q)
	return checkCloneImpl(v, w)
}

// clone calls x.Clone() if possible, otherwise it returns an error.
func clone(x interface{}) (interface{}, error) {
	if x == nil {
		return nil, errors.Errorf("Cannot clone nil reference")
	}
	if !isCloneable(reflect.TypeOf(x)) {
		return nil, errors.Errorf("Input of type %T is not cloneable", x)
	}

	v := reflect.ValueOf(x)
	if clone := v.MethodByName("Clone"); clone.IsValid() {
		// num return values is checked by `isCloneable`
		return clone.Call([]reflect.Value{})[0].Interface(), nil
	}

	panic(fmt.Sprintf("Error when calling %T.Clone() with object %v", x, x))
}

// VerifyClone attemps to recognize improper cloning.
// Initially, this function will clone its input `x` by calling `x.Clone()`,
// where `x` is an instance of a struct (or a reference). Then it attemps to
// detect improper clones by taking the following steps:
// * Run `reflect.DeepEqual` and terminate with an error if it returns false.
//
// Then, for every exported field of `x`:
// * If the field of type `T` is itself is a cloneable, then this value is
//   checked recursively.
// * If the field has kind pointer or slice and if it has a
//   `cloneable:"shallow"` tag, it is checked that the pointer or slice value
//   are the same.
// * If the field has kind array or slice and a `cloneable:"shallowElements"`
//   tag, it is checked that the the array and slice values shallow copies.
//
// Tags attached to inappropriate fields as well as unknown `cloneable` tags
// cause an error. The code was not tested with some possible kinds (e.g.,
// channels, maps, and unsafe pointers) and will immediately panic when seeing
// these types.
func VerifyClone(t *testing.T, x interface{}) {
	if !isCloneable(reflect.TypeOf(x)) {
		t.Errorf("Expected cloneable input, got %v (type %T)", x, x)
	}

	c, err := clone(x)
	if err != nil {
		t.Errorf("Cloning failure: %v", err)
	}
	if err = checkClone(x, c); err != nil {
		t.Error(err)
	}
}
