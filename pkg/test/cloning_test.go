// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import(
	"math/big"
	"reflect"
	"testing"
)


type Cloneable struct {
}

func (Cloneable) Clone() Cloneable {
	return Cloneable{}
}


type CloneableRef struct {
	// this field stops the golang compiler from creating a single global
	// instance of this type that is being referenced everywhere
	identity int
}

func (this *CloneableRef) Clone() *CloneableRef {
	return &CloneableRef{this.identity}
}


type BrokenCloneableRef struct {
}

func (this *BrokenCloneableRef) Clone() *BrokenCloneableRef {
	return this
}


type NotCloneable struct {
}


type NotCloneableInt struct {
}

func (NotCloneableInt) Clone() int {
	return 0
}


type NotCloneableIntRef struct {
}

func (*NotCloneableIntRef) Clone() int {
	return 1
}


type RecursivelyCloneable struct {
	X Cloneable
}

func (this RecursivelyCloneable) Clone() RecursivelyCloneable {
	return RecursivelyCloneable{this.X.Clone()}
}


func Test_isCloneable(t *testing.T) {
	tests := []struct {
		Input interface{}
		Result bool
	}{
		{Cloneable{}, true},
		{CloneableRef{}, true},
		{RecursivelyCloneable{}, true},
		{NotCloneable{}, false},
		{NotCloneableInt{}, false},
		{NotCloneableIntRef{}, false},
		{&Cloneable{}, true},
		{&CloneableRef{}, true},
		{&RecursivelyCloneable{}, true},
		{&NotCloneable{}, false},
		{&NotCloneableInt{}, false},
		{&NotCloneableIntRef{}, false},
		{1, false},
		{1.0, false},
		{[]int{1,2,3}, false},
	}

	for _, test := range tests {
		inputType := reflect.TypeOf(test.Input)
		result := isCloneable(inputType)

		if result != test.Result {
			format := "Expected isCloneable(%T) = %v, got %v"
			t.Errorf(format, test.Input, test.Result, result)
		}
	}
}


func Test_clone(t* testing.T) {
	// this test succeeds even if the return type of `Clone()` is incorrect
	tests := []struct {
		Input interface{}
		CloneShouldSucceed bool
	}{
		{Cloneable{}, true},
		{&CloneableRef{}, true},
		{NotCloneable{}, false},
		{&NotCloneable{}, false},
		{1, false},
		{1.0, false},
	}

	for _, test := range tests {
		x := test.Input
		c, err := clone(test.Input)

		if c == nil && err == nil || c != nil && err != nil {
			format := "Expected one non-nil return value, got clone(%T)=(%v,%v)"
			t.Errorf(format, x, c, err)
		}

		if test.CloneShouldSucceed {
			if c == nil {
				format := "Expected non-nil first return value by clone(%T)"
				t.Errorf(format, x)
			}
			if err != nil {
				format := "Expected nil error message by clone(%T), got %v"
				t.Errorf(format, x, err)
			}
		} else {
			if c != nil {
				format := "Expected nil first return value by clone(%T), got %v"
				t.Errorf(format, x, c)
			}
			if err == nil {
				format := "Expected error message by clone(%T), got nil"
				t.Errorf(format, x)
			}
		}
	}
}



// These structs test if `checkClone` detects improperly cloned field with
// `shallow` tag.

type BrokenShallowClonePtr struct {
	x *CloneableRef `cloneable:"shallow"`
}

func (this BrokenShallowClonePtr) Clone() BrokenShallowClonePtr {
	return BrokenShallowClonePtr{this.x.Clone()}
}


type BrokenShallowCloneSlice struct {
	Xs []int `cloneable:"shallow"`
}

func (this BrokenShallowCloneSlice) Clone() BrokenShallowCloneSlice {
	clone := BrokenShallowCloneSlice{make([]int, len(this.Xs))}

	copy(clone.Xs, this.Xs)

	return clone
}


type BrokenShallowCloneSliceLen struct {
	xs []int `cloneable:"shallow"`
}

func (this BrokenShallowCloneSliceLen) Clone() BrokenShallowCloneSliceLen {
	return BrokenShallowCloneSliceLen{this.xs[:1]}
}


// These struct test if `checkClone` detects improper field clones with
// `shallowSlice` tag.

type BrokenShallowSliceClone struct {
	Xs []int `cloneable:"shallowSlice"`
}

func (this BrokenShallowSliceClone) Clone() (clone BrokenShallowSliceClone) {
	clone = BrokenShallowSliceClone{this.Xs}
	return
}


type BrokenShallowSliceCloneLength struct {
	Xs []int `cloneable:"shallowSlice"`
}

func (this BrokenShallowSliceCloneLength) Clone() (clone BrokenShallowSliceCloneLength) {
	Xs := this.Xs
	clone = BrokenShallowSliceCloneLength{make([]int, len(Xs)+1)}

	copy(clone.Xs, Xs)

	return
}


type BrokenDeepClone struct {
	Xs []*big.Float
}

func (this BrokenDeepClone) Clone() (clone BrokenDeepClone) {
	Xs := this.Xs
	clone = BrokenDeepClone{make([]*big.Float, len(Xs)-1)}

	for i := 0; i < len(Xs)-1; i++ {
		clone.Xs[i] = Xs[i]
	}

	return
}



type UnknownTag struct {
	xs []int `cloneable:"thisIsNotATag"`
}

func (this UnknownTag) Clone() UnknownTag {
	return UnknownTag{this.xs}
}



func Test_checkClone(t *testing.T) {
	tests := []struct {
		Input interface{}
		ExpectProperClone bool
	}{
		{Cloneable{}, true},
		{&CloneableRef{}, true},
		{&BrokenCloneableRef{}, false},
		{RecursivelyCloneable{}, true},
		{BrokenShallowClonePtr{&CloneableRef{1}}, false},
		{BrokenShallowCloneSlice{[]int{1,2,3}}, false},
		{BrokenShallowCloneSliceLen{[]int{1,2,3}}, false},
		{BrokenShallowSliceClone{[]int{1,2,3}}, false},
		{BrokenShallowSliceCloneLength{[]int{1,2,3}}, false},
		{BrokenDeepClone{[]*big.Float{big.NewFloat(1), big.NewFloat(2)}},false},
		{UnknownTag{[]int{1}}, false},
	}

	for _, test := range tests {
		x := test.Input
		c, err := clone(x)

		if err != nil {
			t.Fatalf("BUG: clone error: %v", err)
		}

		err = checkClone(x, c)

		if test.ExpectProperClone && err != nil {
			format := "Expected checkClone(%T) to return nil, got error '%v'"
			t.Errorf(format, x, err)
		}

		if !test.ExpectProperClone && err == nil {
			t.Errorf("Expected checkClone(%T) to return a non-nil value", x)
		}
	}
}
