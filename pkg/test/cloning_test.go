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


type BrokenCloneablePtr struct {
	x *int
}

func (this BrokenCloneablePtr) Clone() BrokenCloneablePtr {
	return BrokenCloneablePtr{}
}


type NotCloneable struct {
}


type NotCloneableNumArgsIn struct {
}

func (NotCloneableNumArgsIn) Clone(NotCloneableNumArgsIn) NotCloneableNumArgsIn{
	return NotCloneableNumArgsIn{}
}


type NotCloneableNumArgsOut struct {
}

func (NotCloneableNumArgsOut) Clone() (NotCloneableNumArgsOut,NotCloneableNumArgsOut) {
	return NotCloneableNumArgsOut{}, NotCloneableNumArgsOut{}
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

type RecursivelyCloneableRef struct {
	x *RecursivelyCloneableRef
}

func (this *RecursivelyCloneableRef) Clone() *RecursivelyCloneableRef {
	return &RecursivelyCloneableRef{this.x.Clone()}
}


func Test_isCloneable(t *testing.T) {
	tests := []struct {
		Input interface{}
		Result bool
	}{
		{Cloneable{}, true},
		{CloneableRef{}, true},
		{RecursivelyCloneable{}, true},
		{RecursivelyCloneableRef{}, true},
		{RecursivelyCloneableRef{&RecursivelyCloneableRef{}}, true},
		{NotCloneable{}, false},
		{NotCloneableNumArgsIn{}, false},
		{NotCloneableNumArgsOut{}, false},
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
		{[]int{1}, false},
		{nil, false},
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


// detect broken clone detection for nested cloneable structures

type BrokenCloneableNestedInner struct {
	x *int
}

func (this BrokenCloneableNestedInner) Clone() BrokenCloneableNestedInner {
	return BrokenCloneableNestedInner{this.x}
}

type BrokenCloneableNested struct {
	inner BrokenCloneableNestedInner
}

func (this BrokenCloneableNested) Clone() BrokenCloneableNested {
	return BrokenCloneableNested{this.inner.Clone()}
}


// detect broken clone detection for arrays

type BrokenCloneableNestedArray struct {
	inner [1]BrokenCloneableNestedInner
}

func (this BrokenCloneableNestedArray) Clone() BrokenCloneableNestedArray {
	return BrokenCloneableNestedArray{
		[1]BrokenCloneableNestedInner{this.inner[0].Clone()}}
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
// `shallowElements` tag.

type BrokenShallowElementsClone struct {
	Xs []int `cloneable:"shallowElements"`
}

func (this BrokenShallowElementsClone) Clone() (clone BrokenShallowElementsClone) {
	clone = BrokenShallowElementsClone{this.Xs}
	return
}


type BrokenShallowElementsCloneLength struct {
	Xs []int `cloneable:"shallowElements"`
}

func (this BrokenShallowElementsCloneLength) Clone() (clone BrokenShallowElementsCloneLength) {
	Xs := this.Xs
	clone = BrokenShallowElementsCloneLength{make([]int, len(Xs)+1)}

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



type MisplacedTagShallow struct {
	x uint `cloneable:"shallow"`
}

func (this MisplacedTagShallow) Clone() MisplacedTagShallow {
	return MisplacedTagShallow{this.x}
}


type MisplacedTagShallowElements struct {
	x *CloneableRef `cloneable:"shallowElements"`
}

func (this MisplacedTagShallowElements) Clone() MisplacedTagShallowElements {
	return MisplacedTagShallowElements{this.x.Clone()}
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
		{&BrokenCloneableRef{}, false},
		{BrokenCloneablePtr{new(int)}, false},
		{BrokenCloneableNested{BrokenCloneableNestedInner{new(int)}}, false},
		{BrokenCloneableNestedArray{
			[1]BrokenCloneableNestedInner{BrokenCloneableNestedInner{new(int)}}},
			false},
		{RecursivelyCloneable{}, true},
		{BrokenShallowClonePtr{&CloneableRef{1}}, false},
		{BrokenShallowCloneSlice{[]int{1,2,3}}, false},
		{BrokenShallowCloneSliceLen{[]int{1,2,3}}, false},
		{BrokenShallowElementsClone{[]int{1,2,3}}, false},
		{BrokenShallowElementsCloneLength{[]int{1,2,3}}, false},
		{BrokenDeepClone{[]*big.Float{big.NewFloat(1), big.NewFloat(2)}},false},
		{MisplacedTagShallow{0}, false},
		{MisplacedTagShallowElements{&CloneableRef{0}}, false},
		{UnknownTag{[]int{1}}, false},
	}

	for _, test := range tests {
		x := test.Input
		c, err := clone(x)

		if err != nil {
			t.Fatalf("BUG: clone error: %v", err)
		}

		err = checkClone(x, c)

		if err != nil {
			println(err.Error())
		}

		if test.ExpectProperClone && err != nil {
			format := "Expected checkClone(%T) to return nil, got error '%v'"
			t.Errorf(format, x, err)
		}

		if !test.ExpectProperClone && err == nil {
			t.Errorf("Expected checkClone(%T) to return a non-nil value", x)
		}
	}
}



// the code below tests `checkClone` with a more complex type.

// This is a linked list node for a functional programming language meaning
// only the preceeding nodes change. Below, y was modified to become y':
//
// x  -> y  -> z
// x' -> y' ---^
type ListNode struct {
	prev *ListNode
	next *ListNode `cloneable:"shallow"`
	integer uint
	xs []*big.Float
	ys []*big.Float `cloneable:"shallowElements"`
}

func (this *ListNode) ShallowClone() *ListNode {
	clone := &ListNode{
		nil,
		this.next,
		this.integer,
		make([]*big.Float, len(this.xs)),
		make([]*big.Float, len(this.ys)),
	}

	if this.prev != nil {
		clone.prev = this.prev.Clone()
	}

	copy(clone.xs, this.xs)
	copy(clone.ys, this.ys)

	return clone
}

func (this *ListNode) Clone() *ListNode {
	clone := this.ShallowClone()

	for i, x := range this.xs {
		y := big.NewFloat(0)
		y.Copy(x)
		clone.xs[i] = y
	}

	return clone
}



type SelfContained struct {
	xs []SelfContained
	alwaysNil []SelfContained
}

func (this SelfContained) Clone() SelfContained {
	n := len(this.xs)
	clone := SelfContained{make([]SelfContained, n), nil}

	if n != 0 {
		for i, _ := range this.xs {
			clone.xs[i] = this.xs[i].Clone()
		}
	}

	return clone
}



type HasArray struct {
	xs [2]CloneableRef
	ys [1]*big.Float `cloneable:"shallowElements"`
	zs [1]int `cloneable:"shallowElements"`
}

func (this HasArray) Clone() HasArray {
	return HasArray{
		[2]CloneableRef{*this.xs[0].Clone(), *this.xs[1].Clone()},
		[1]*big.Float{this.ys[0]},
		[1]int{this.zs[0]},
	}
}



// "manually" because the clones are computed individually
func Test_checkCloneManually(t *testing.T) {
	p0 := ListNode{
		nil, nil, 1, []*big.Float{big.NewFloat(1)}, []*big.Float{big.NewFloat(-1)},
	}
	p1 := ListNode{
		&p0, nil, 2, []*big.Float{big.NewFloat(2)}, []*big.Float{big.NewFloat(-2)},
	}
	p2 := ListNode{
		&p1, nil, 3, []*big.Float{big.NewFloat(3)}, []*big.Float{big.NewFloat(-3)},
	}

	p0.next = &p1
	p1.next = &p2

	ss := SelfContained{[]SelfContained{
		SelfContained{[]SelfContained{}, nil},
		SelfContained{[]SelfContained{}, nil}}, nil}

	arr0 := HasArray{}
	arr0.ys[0] = big.NewFloat(0)
	arr1 := arr0.Clone()
	arr1.ys[0] = big.NewFloat(0)
	arr1.ys[0].Copy(arr0.ys[0])

	rcr := RecursivelyCloneableRef{}
	rcr0 := RecursivelyCloneableRef{&RecursivelyCloneableRef{&rcr}}
	rcr1 := RecursivelyCloneableRef{&RecursivelyCloneableRef{&rcr}}

	tests := []struct {
		Original interface{}
		Clone interface{}
		IsProperClone bool
	}{
		{&p0, p0.Clone(), true},
		{ss, ss.Clone(), true},
		{arr0, arr0.Clone(), true},
		{nil, nil, false},
		{NotCloneable{}, NotCloneable{}, false},
		{p0, 1, false},
		{p0, &p0, false},
		{&p0, p0, false},
		{p0, p0.Clone(), false},
		{&p0, p0.ShallowClone(), false},
		{ss, ss, false},
		{arr0, arr1, false},
		{rcr0, rcr1, false},
	}

	for _, test := range tests {
		err := checkClone(test.Original, test.Clone)

		if err != nil {
			println(err.Error())
		}

		if test.IsProperClone && err != nil {
			format := "Expected checkClone(%T) to return nil, got error '%v'"
			t.Errorf(format, test.Original, err)
		}

		if !test.IsProperClone && err == nil {
			t.Errorf("Expected checkClone(%T) to return a non-nil value", test.Original)
		}
	}
}
