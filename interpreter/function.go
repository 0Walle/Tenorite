package interpreter

import (
    "fmt"
)

type CodeObj struct {
    Code          []uint16
    Consts        []Receiver
    // Module        *Module
    Arity         uint16
    LocalSize     uint16
    UpvalueCount  uint16
    Name          string

    BaseLine      int
    Lines         []int
    CoVarnames    []string
    DebugMap      map[uint16]string
}

type Upvalue struct {
    Value   *Receiver
    Slot    uint16
    Closed  Receiver
    Next    *Upvalue
}

type Closure struct {
    CodeObj       *CodeObj
    Upvalues      []*Upvalue
}

func (c Closure) String() string {
    return fmt.Sprintf("{%s}", c.CodeObj.Name)
}

type Primitive struct {
    Call func(*TenoriteVM, []Receiver) Receiver
}

func (_ CodeObj) GetMethod(_ Symbol) Receiver { return nil }
func (_ CodeObj) Type(_ Receiver) bool { return false }