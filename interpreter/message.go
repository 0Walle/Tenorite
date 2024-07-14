package interpreter

import (
	"fmt"
)

type Message struct {
	Symbol Symbol
	Ranks  []int
}

func CallUnsafe(vm *TenoriteVM, recv Receiver, name string, args []Receiver) Receiver {
	r, e := Call(vm, Message { vm.Symbol(name), nil }, append([]Receiver{recv}, args...) )
	if e != nil {
		panic(e)
	}
	return r
}

func allZero(ranks []int) bool {
	for _, i := range ranks {
		if i > 0 { return false }
	}
	return true
}

func Call(vm *TenoriteVM, msg Message, args []Receiver) (Receiver, error) {
	// _, isList := args[0].(List)
	// if isList && (msg.Ranks == nil || msg.Ranks[0] == 0) {
	if IsCollection(args[0]) && (msg.Ranks == nil || allZero(msg.Ranks)) {
		method := args[0].GetMethod(msg.Symbol)
		if method == nil {
			msg.Ranks[0] = 1

			// return Run(vm, method, args)
		}
		// fmt.Printf("Ranks %s %v\n", msg.Symbol, msg.Ranks)
		return CallRec_(vm, 0, msg, args)
	}

	return CallRec_(vm, 0, msg, args)
}

func CallRec_(vm *TenoriteVM, depth int, msg Message, args []Receiver) (Receiver, error) {
	arity := len(msg.Ranks)

	var toZip []int
	for i := 0; i < arity; i++ {
		if msg.Ranks[i]-depth == 1 && IsCollection(args[i]) {
			toZip = append(toZip, i)
		}
	}

	if toZip == nil {
		method := args[0].GetMethod(msg.Symbol)
		if method == nil {
			return nil, fmt.Errorf("Invalid method #%s for %v. Ranks %v", vm.SymbolStore[msg.Symbol], args[0], msg.Ranks)
		}
		return Run(vm, method, args)
	}
	

	size := Size(args[toZip[0]])
	for _, i := range toZip {
		if Size(args[i]) != size {
			return nil, fmt.Errorf("Differing sizes")
		}
	}

	// fmt.Printf("ziping %v Ranks %s %v-%d\n", toZip, msg.Symbol, msg.Ranks, depth)

	var result = make([]Receiver, size)
	for i := 0; i < size; i++ {
		var newArgs = make([]Receiver, arity)
		for j := 0; j < arity; j++ {
			if msg.Ranks[j]-depth == 1 {
				newArgs[j] = GetAt(args[j], i)
			} else {
				newArgs[j] = args[j]
			}
		}

		r, err := CallRec_(vm, depth+1, msg, newArgs)
		if err != nil { return nil, err }
		result[i] = r
	}

	if msg.Ranks[0] == 1 {
		tbl, isTable := args[0].(Table)
		if isTable {
			collected := Table { tbl.Keys, result }
			return collected, nil
		}

	}
	collected := List { result }

	return collected, nil
}