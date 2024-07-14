package interpreter

import (
    "regexp"
    "fmt"
)

type Regex struct {
    Regex *regexp.Regexp
}

func RegexNew(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateString(vm, args[1], "Regex") { return nil }
    patt := string(args[1].(String))
    regex, err := regexp.Compile(patt)
    if err != nil {
        vm.Error = err
        return nil
    }
    return Regex { regex }
}

func StringFindRegex(vm *TenoriteVM, args []Receiver) Receiver {
    self := args[0].(String)
    patt, ok := args[1].(Regex)
    if !ok {
        vm.Error = fmt.Errorf("Argument must be regex.")
        return nil
    }
    size := patt.Regex.NumSubexp()+1
    indexes := patt.Regex.FindStringSubmatchIndex(string(self))
    if indexes == nil {
        size = 0
    }
    groups := make([]Receiver, size)
    spans := make([]Receiver, size)

    if indexes != nil {
        for i := 0; i < size; i++ {
            if indexes[i*2] < 0 {
                groups[i] = String("")
                spans[i] = NONE
                continue
            }
            groups[i] = String(self[indexes[i*2]:indexes[i*2+1]])
            spans[i] = Range { Number(indexes[i*2]), Number(indexes[i*2+1]) }
        }
    }
    objmap := make(map[Symbol]Receiver, 4)
    objmap[vm.Symbol("groups")] = List { groups }
    objmap[vm.Symbol("spans")] = List { spans }
    objmap[vm.Symbol("matched")] = toBool(indexes != nil)
    objmap[vm.Symbol("subject")] = self
    return Object { []*Namespace{RegexResultsNs}, objmap }
}

func StringFindRegexStart(vm *TenoriteVM, args []Receiver) Receiver {
    self_ := args[0].(String)
    patt, ok := args[1].(Regex)
    if !ok {
        vm.Error = fmt.Errorf("Argument must be regex.")
        return nil
    }
    if !validateNumber(vm, args[2], "Argument") { return nil }
    start := int(args[2].(Number))
    self := string(self_)[start:]

    size := patt.Regex.NumSubexp()+1
    indexes := patt.Regex.FindStringSubmatchIndex(self)
    if indexes == nil {
        size = 0
    }
    groups := make([]Receiver, size)
    spans := make([]Receiver, size)

    if indexes != nil {
        for i := 0; i < size; i++ {
            if indexes[i*2] < 0 {
                groups[i] = String("")
                spans[i] = NONE
                continue
            }
            groups[i] = String(self[indexes[i*2]:indexes[i*2+1]])
            spans[i] = Range { Number(start + indexes[i*2]), Number(start + indexes[i*2+1]) }
        }
    }
    objmap := make(map[Symbol]Receiver, 4)
    objmap[vm.Symbol("groups")] = List { groups }
    objmap[vm.Symbol("spans")] = List { spans }
    objmap[vm.Symbol("matched")] = toBool(indexes != nil)
    objmap[vm.Symbol("subject")] = self_
    return Object { []*Namespace{RegexResultsNs}, objmap }
}
