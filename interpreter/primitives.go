package interpreter

import (
    "hash/fnv"
    "fmt"
    // "strings"
)

type Receiver interface {
    GetMethod(Symbol) Receiver
    Type(Receiver) bool
}

func Size(r Receiver) int {
    switch recv := r.(type) {
    case List: return len(recv.List)
    case Table: return len(recv.Keys)
    }
    return 1
}

func IsCollection(r Receiver) bool {
    switch r.(type) {
    case List: return true
    case Table: return true
    }
    return false
}

func GetAt(r Receiver, index int) Receiver {
    switch recv := r.(type) {
    case List: return recv.List[index]
    case Table: return recv.Values[index]
    }
    return r
}

// == Basic ==

type True struct {}
type False struct {}
type None struct {}

var TRUE = True{}
var FALSE = False{}
var NONE = None{}

type String string
type Number float64


func (_ True) String() string { return "True" }
func (_ False) String() string { return "False" }
func (_ None) String() string { return "None" }

// == Collections ==

type List struct {
    List []Receiver
}

type Table struct {
    Keys     []Receiver
    Values   []Receiver
}

type Pair struct {
    First   Receiver
    Second  Receiver
}

type Range struct {
    From  Number
    To    Number
}

// == Symbol ==

var hasher = fnv.New32a()

type Symbol uint16

func makeSymbol(name string) Symbol {
    hasher.Reset()
    hasher.Write([]byte(name))
    return Symbol(hasher.Sum32())
}

func (sym Symbol) String() string {
    return fmt.Sprintf("<#%d>", sym)
}

var (
    SYM_ITERATE = makeSymbol("iterate:")
    SYM_VALUE = makeSymbol("value:")
)

// == Namespace ==

type Namespace struct {
    Table   map[Symbol]Receiver
    Name    string
    Static  *Namespace
    Locked  bool
}

func NewNamespace(name string) *Namespace {
    return &Namespace{ make(map[Symbol]Receiver, 0), name, nil, false }
}

func (ns *Namespace) Set(name Symbol, value Receiver) {
    ns.Table[name] = value
}

func (ns *Namespace) Get(name Symbol) Receiver {
    return ns.Table[name]
}

func (ns *Namespace) String() string {
    return "<"+ns.Name+">"
}

// == Tuple ==

type Object struct {
    Roles  []*Namespace
    Table  map[Symbol]Receiver
}

// == Foreign ==

// type Foreign struct {
//     Foreign any
// }

// func (_ Foreign) Namespace() *Namespace { return nil }

var ObjectNs = NewNamespace("Object")
var BoolNs = NewNamespace("Bool")
var StringNs = NewNamespace("String")
var NumberNs = NewNamespace("Number")
var FunctionNs = NewNamespace("Function")
var ListNs = NewNamespace("List")
var TableNs = NewNamespace("Table")
var RangeNs = NewNamespace("Range")
var NamespaceNs = NewNamespace("Namespace")
var SymbolNs = NewNamespace("Symbol")
var PairNs = NewNamespace("Pair")

var RegexNs = NewNamespace("Regex")
var RegexResultsNs = NewNamespace("RegexResults")

var EqNs = NewNamespace("Eq")
var OrdNs = NewNamespace("Ord")

func (_ None) GetMethod(sym Symbol) Receiver {
    return ObjectNs.Get(sym)
}
func (_ True) GetMethod(sym Symbol) (meth Receiver) {
    meth = BoolNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}
func (_ False) GetMethod(sym Symbol) (meth Receiver) {
    meth = BoolNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}
func (_ String) GetMethod(sym Symbol) (meth Receiver) {
    meth = StringNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}
func (_ Number) GetMethod(sym Symbol) (meth Receiver) {
    meth = NumberNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}
func (_ Symbol) GetMethod(sym Symbol) Receiver {
    return ObjectNs.Get(sym)
}
func (_ Range) GetMethod(sym Symbol) (meth Receiver) {
    meth = RangeNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}
func (_ Pair) GetMethod(sym Symbol) (meth Receiver) {
    meth = PairNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}
func (_ List) GetMethod(sym Symbol) (meth Receiver) {
    meth = ListNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}
func (_ Table) GetMethod(sym Symbol) (meth Receiver) {
    meth = TableNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}
func (self *Namespace) GetMethod(sym Symbol) Receiver {
    if self.Static != nil {
        meth := self.Static.Get(sym)
        if meth != nil { return meth }
    }

    meth := NamespaceNs.Get(sym); if meth != nil { return meth }

    return ObjectNs.Get(sym)
}
func (self Object) GetMethod(sym Symbol) Receiver {
    for _, role := range self.Roles {
        meth := role.Get(sym)
        if meth != nil {
            return meth
        }
    }
    return ObjectNs.Get(sym)
}
func (_ *Closure) GetMethod(sym Symbol) (meth Receiver) {
    meth = FunctionNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}
func (_ Primitive) GetMethod(sym Symbol) (meth Receiver) {
    meth = FunctionNs.Get(sym); if meth != nil { return }
    return ObjectNs.Get(sym)
}

func (_ Regex) GetMethod(sym Symbol) (meth Receiver) {
    return ObjectNs.Get(sym)
}










func (_ None) Type(r Receiver) bool { return r == ObjectNs }
func (_ True) Type(r Receiver) bool { return r == BoolNs || r == ObjectNs }
func (_ False) Type(r Receiver) bool { return r == BoolNs || r == ObjectNs }
func (_ String) Type(r Receiver) bool { return r == StringNs || r == ObjectNs }
func (_ Number) Type(r Receiver) bool { return r == NumberNs || r == ObjectNs }
func (_ Symbol) Type(r Receiver) bool { return r == SymbolNs || r == ObjectNs }
func (_ Range) Type(r Receiver) bool { return r == RangeNs || r == ObjectNs }
func (_ Pair) Type(r Receiver) bool { return r == PairNs || r == ObjectNs }
func (_ List) Type(r Receiver) bool { return r == ListNs || r == ObjectNs }
func (_ Table) Type(r Receiver) bool { return r == TableNs || r == ObjectNs }
func (_ *Namespace) Type(r Receiver) bool { return r == NamespaceNs || r == ObjectNs }
func (_ *Closure) Type(r Receiver) bool { return r == FunctionNs || r == ObjectNs }
func (_ Primitive) Type(r Receiver) bool { return r == FunctionNs || r == ObjectNs }
func (self Object) Type(r Receiver) bool {
    for _, role := range self.Roles {
        if role == r { return true }
    }
    return false
}

func (_ Regex) Type(r Receiver) bool { return r == RegexNs || r == ObjectNs }
