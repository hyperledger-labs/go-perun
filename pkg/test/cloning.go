// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

// VerifyClone attempts to recognize improper cloning.
// Initially, this function will clone its input `x` by calling `x.Clone()`,
// where `x` is an instance of a struct (or a reference). Then it attempts to
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

// clone calls x.Clone() if possible, otherwise it returns an error.
func clone(x interface{}) (interface{}, error) {
	if x == nil {
		return nil, errors.Errorf("cannot clone nil reference")
	}
	if !isCloneable(reflect.TypeOf(x)) {
		return nil, errors.Errorf("input of type %T is not cloneable", x)
	}

	v := reflect.ValueOf(x)
	if clone := v.MethodByName("Clone"); clone.IsValid() {
		// num return values is checked by `isCloneable`
		return clone.Call([]reflect.Value{})[0].Interface(), nil
	}

	panic(fmt.Sprintf("Error when calling %T.Clone() with object %v", x, x))
}

// checkClone checks if the two provided values could be clones. If they are
// not, an error is returned.
//
// checkClone initially calls reflect.DeepEqual and then checkCloneImpl tests
// recursively if the provided values could be clones.
func checkClone(p, q interface{}) error {
	if !reflect.DeepEqual(p, q) {
		return errors.New("proper clones must be deeply equal")
	}
	if p == nil || q == nil {
		return errors.Errorf("input must not be nil, got %v, %v", p, q)
	}

	tP := reflect.TypeOf(p)
	tQ := reflect.TypeOf(q)
	if tP != tQ || !isCloneable(tP) {
		return errors.Errorf("input must be cloneable, type %v is not", tP)
	}

	v := reflect.ValueOf(p)
	w := reflect.ValueOf(q)
	return checkCloneImpl(v, w)
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
	if err := validateInput(v, w); err != nil {
		return err
	}

	// get struct type
	baseType := v.Type()
	if baseType.Kind() == reflect.Ptr {
		v = v.Elem()
		w = w.Elem()
		baseType = baseType.Elem()
	}

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
		if hasTag {
			if err := checkTags(f, baseType, tag, kind); err != nil {
				return errors.WithMessage(err, "wrong or missing tags")
			}
		}

		// check actual field contents
		if valid := checkInterface(&kind, &left, &right); !valid {
			continue
		}

		if err := checkPtrOrSlice(f, tag, hasTag, kind, left, right, baseType); err != nil {
			return err
		}

		if err := checkArrayOrSlice(f, tag, hasTag, kind, left, right, baseType); err != nil {
			return err
		}
	}

	return nil
}

func validateInput(v, w reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		if v.Pointer() == 0 && w.Pointer() == 0 {
			panic("BUG: checkCloneImpl got nil inputs")
		}
		if v.Pointer() == w.Pointer() {
			return errors.New("both arguments reference the same structure")
		}
		if v.Elem().Kind() != reflect.Struct {
			panic("BUG: expected reference to struct, got reference to reference")
		}
	}

	return nil
}

func checkTags(f reflect.StructField, t reflect.Type, tag string, kind reflect.Kind) error {
	// find unknown and misplaced tags
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
	return nil
}

func checkInterface(kind *reflect.Kind, left, right *reflect.Value) bool {
	if *kind == reflect.Interface {
		if left.IsZero() && right.IsZero() {
			return false
		}

		if left.IsZero() != right.IsZero() {
			// reflect.DeepEqual() should detect this case
			panic(fmt.Sprintf(
				"Expected both interfaces to be zero or both non-zero, got %v, %v",
				left.InterfaceData(), right.InterfaceData()))
		}

		*left = left.Elem()
		*right = right.Elem()
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

		*kind = left.Type().Kind()
	}
	return true
}

func checkPtrOrSlice(f reflect.StructField, tag string, hasTag bool, kind reflect.Kind, left, right reflect.Value, t reflect.Type) error {
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
			return errors.WithMessagef(err, "error in cloneable field %v.%s", t, f.Name)
		}
	}
	return nil
}

func checkArrayOrSlice(f reflect.StructField, tag string, hasTag bool, kind reflect.Kind, left, right reflect.Value, t reflect.Type) error {
	if kind == reflect.Array || kind == reflect.Slice {
		n := left.Len()
		for j := 0; j < n; j++ {
			kindJ := left.Index(j).Kind()
			if err := checkPtrOrSliceElem(f, kindJ, tag, hasTag, j, left, right, t); err != nil {
				return err
			}
		}
	} else if kind == reflect.Struct && isCloneable(f.Type) {
		err := checkCloneImpl(left, right)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkPtrOrSliceElem(f reflect.StructField, kindJ reflect.Kind, tag string, hasTag bool, j int, left, right reflect.Value, t reflect.Type) error {
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
		return errors.WithMessagef(err, "error in cloneable element %v.%s[%d]", t, f.Name, j)
	}
	return nil
}
