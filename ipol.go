package main

import (
	"strings"
)

type FuncIpolFn func(...string) (string, error)

type FuncIpol struct {
	data       map[string]FuncIpolFn
	leftDelim  string
	rightDelim string
}

func newFuncIpol(leftDelim, rightDelim string) *FuncIpol {
	return &FuncIpol{
		data:       make(map[string]FuncIpolFn),
		leftDelim:  leftDelim,
		rightDelim: rightDelim,
	}
}

func (f *FuncIpol) RegisterFn(name string, fn FuncIpolFn) {
	f.data[strings.Trim(name, " ")] = fn
}

func (f *FuncIpol) Eval(s string) (string, error) {
	partsA := strings.Split(s, f.leftDelim)

	for i, a := range partsA {
		if !strings.Contains(a, f.rightDelim) {
			continue
		}

		partsB := strings.Split(a, f.rightDelim)
		if len(partsB) == 0 {
			continue
		}

		trimmed := strings.TrimSpace(partsB[0])
		partsC := strings.Split(trimmed, "(")
		if len(partsC) < 2 {
			continue
		}

		if popStr(&partsC[len(partsC)-1]) != ")" {
			continue
		}

		var (
			name = partsC[0]
			args = strings.Split(partsC[1], ",")
		)

		fn, ok := f.data[name]
		if !ok {
			continue
		}

		s, err := fn(args...)
		if err != nil {
			continue
		}
		partsB[0] = s

		partsA[i] = strings.Join(partsB, "")
	}

	return strings.Join(partsA, ""), nil
}
