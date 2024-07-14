package interpreter

import (
    "fmt"
    "strings"
    "strconv"
    "math"
)


func validateString(vm *TenoriteVM, value Receiver, name string) bool {
    _, ok := value.(String)
    vm.Error = fmt.Errorf("%s must be string.", name)
    return ok
}

func validateNumber(vm *TenoriteVM, value Receiver, name string) bool {
    _, ok := value.(Number)
    vm.Error = fmt.Errorf("%s must be number.", name)
    return ok
}

func validateList(vm *TenoriteVM, value Receiver, name string) bool {
    _, ok := value.(List)
    vm.Error = fmt.Errorf("%s must be a list.", name)
    return ok
}

func toBool(b bool) Receiver {
    if b { return TRUE }
    return FALSE
}

func isFalsey(r Receiver) bool {
    if _, isFalse := r.(False); isFalse { return true }
    if _, isNone := r.(None); isNone { return true }
    return false
}

func toString(vm *TenoriteVM, r Receiver) string {
    switch recv := r.(type) {
    case None: return "None"
    case True: return "True"
    case False: return "False"
    case String: return string(recv)
    case Symbol: return fmt.Sprintf("#%s", vm.SymbolStore[recv])
    case Number: return fmt.Sprintf("%g", float64(recv))
    case Pair: return fmt.Sprintf("%s => %s", toDebugString(vm, recv.First), toDebugString(vm, recv.Second))
    case Range: return fmt.Sprintf("%g;%g", float64(recv.From), float64(recv.To))
    case *Namespace: return fmt.Sprintf("<%s>", recv.Name)
    case *Closure: return fmt.Sprintf("<Function>")
    case Primitive: return fmt.Sprintf("<Function>")
    case Regex: return fmt.Sprintf("#'%s'", recv.Regex.String())
    case Object:
        if len(recv.Roles) == 0 {
            return "<Object>"
        }

        name := recv.Roles[len(recv.Roles)-1].Name
        return fmt.Sprintf("<object %s>", name)
    default: return "<Object>"
    }
}

func toDebugString(vm *TenoriteVM, r Receiver) string {
    switch recv := r.(type) {
    case String: return fmt.Sprintf("%#v", string(recv))
    default: return toString(vm, r)
    }
}

func sameObj(a, b Receiver) bool {
    switch a.(type) {
    case List:
        return false
    default:
        return a == b
    }
}

func formatingSpec(fmt_ string) (fill rune, right bool, width int, rest string) {
    fmt := []rune(fmt_)
    fill = ' '
    widthStart := 0
    if len(fmt) > 2 && fmt[1] == '<' { fill = fmt[0]; widthStart = 2; goto Width }
    if len(fmt) > 2 && fmt[1] == '>' { fill = fmt[0]; widthStart = 2; right = true; goto Width }
    if len(fmt) > 1 && fmt[0] == '<' { widthStart = 1; goto Width }
    if len(fmt) > 1 && fmt[0] == '>' { widthStart = 1; right = true; goto Width }
    Width:
    fmt_ = string(fmt[widthStart:])
    widthEnd := strings.LastIndexAny(fmt_, "0123456789")
    if widthEnd == -1 {
        rest = fmt_
        width = 0
        return
    }
    w, _ := strconv.ParseInt(fmt_[:widthEnd+1], 10, 32)
    width = int(w)
    rest = fmt_[widthEnd+1:]
    return
}
func padStr(fill rune, right bool, width int, str string) string {
    r := width-len([]rune(str))
    if r <= 0 { return str }
    if right {
        return strings.Repeat(string(fill), r)+str
    }
    return str+strings.Repeat(string(fill), r)
}

// ============ Object ============

func ObjSame(vm *TenoriteVM, args []Receiver) Receiver {
    return toBool(sameObj(args[0], args[1]))
}

func ObjNotSame(vm *TenoriteVM, args []Receiver) Receiver {
    return toBool(!sameObj(args[0], args[1]))
}

func ObjPair(vm *TenoriteVM, args []Receiver) Receiver {
    return Pair { args[0], args[1] }
}

func FormatPadding(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateString(vm, args[2], "Format") { return nil }
    format := args[2].(String)
    fill, right, w, _ := formatingSpec(string(format))
    return String(padStr(fill, right, w, toString(vm, args[1])))
}

func ObjString(vm *TenoriteVM, args []Receiver) Receiver {
    return String(toString(vm, args[0]))
}

func NamespaceMake(vm *TenoriteVM, args []Receiver) Receiver {
    ns := args[0].(*Namespace)
    var obj Object
    obj.Table = make(map[Symbol]Receiver)
    obj.Roles = []*Namespace{ ns }

    constructor, ok := args[1].(*Closure)
    if !ok {
        vm.Error = fmt.Errorf("Argument must be a function")
        return nil
    }

    _, err := RunClosure(vm, constructor, []Receiver { obj })
    if err != nil {
        vm.Error = err
        return nil
    }

    return obj
}

// ============ Bool ============

func BoolNot(vm *TenoriteVM, args []Receiver) Receiver {
    if _, isFalse := args[0].(False); isFalse {
        return TRUE
    }
    return FALSE
}
func BoolString(vm *TenoriteVM, args []Receiver) Receiver {
    if _, isFalse := args[0].(False); isFalse {
        return String("False")
    }
    return String("True")
}

// ============ Number ============

func NumAdd(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return args[0].(Number) + args[1].(Number) }
func NumSubtract(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return args[0].(Number) - args[1].(Number) }
func NumMultiply(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return args[0].(Number) * args[1].(Number) }
func NumDivide(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return args[0].(Number) / args[1].(Number) }
func NumModulo(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return Number(int(args[0].(Number)) % int(args[1].(Number))) }
func NumPower(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return Number(math.Pow(float64(args[0].(Number)), float64(args[1].(Number)))) }
func NumMin(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return Number(math.Min(float64(args[0].(Number)), float64(args[1].(Number)))) }
func NumMax(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return Number(math.Max(float64(args[0].(Number)), float64(args[1].(Number)))) }
func NumGt(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return toBool(args[0].(Number) > args[1].(Number)) }
func NumLt(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return toBool(args[0].(Number) < args[1].(Number)) }
func NumGe(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return toBool(args[0].(Number) >= args[1].(Number)) }
func NumLe(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    return toBool(args[0].(Number) <= args[1].(Number)) }

func NumRange(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Right side") { return nil }
    from := args[0].(Number)
    to := args[1].(Number)
    return Range { from, to }
}

func NumString(vm *TenoriteVM, args []Receiver) Receiver {
    return String(fmt.Sprintf("%g", float64(args[0].(Number))))
}

// ============ String ============

func StringGt(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateString(vm, args[1], "Right side") { return nil }
    return toBool(args[0].(String) > args[1].(String)) }
func StringLt(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateString(vm, args[1], "Right side") { return nil }
    return toBool(args[0].(String) < args[1].(String)) }
func StringGe(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateString(vm, args[1], "Right side") { return nil }
    return toBool(args[0].(String) >= args[1].(String)) }
func StringLe(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateString(vm, args[1], "Right side") { return nil }
    return toBool(args[0].(String) <= args[1].(String)) }

func StringLen(vm *TenoriteVM, args []Receiver) Receiver {
    str := args[0].(String)
    return Number(len(str))
}

func StringAt_(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Argument") { return nil }
    str := args[0].(String)
    index := int(args[1].(Number))

    return String(str[index])
}

func StringUpper(vm *TenoriteVM, args []Receiver) Receiver {
    return String(strings.ToUpper(string(args[0].(String))))
}
func StringLower(vm *TenoriteVM, args []Receiver) Receiver {
    return String(strings.ToLower(string(args[0].(String))))
}
func StringTrim(vm *TenoriteVM, args []Receiver) Receiver {
    return String(strings.TrimSpace(string(args[0].(String))))
}
func StringTrimLeft(vm *TenoriteVM, args []Receiver) Receiver {
    return String(strings.TrimLeft(string(args[0].(String)), " "))
}
func StringTrimRight(vm *TenoriteVM, args []Receiver) Receiver {
    return String(strings.TrimRight(string(args[0].(String)), " "))
}

func StringExplode(vm *TenoriteVM, args []Receiver) Receiver {
    var l List
    for _, str := range strings.Split(string(args[0].(String)), "") {
        l.List = append(l.List, String(str))
    }
    return l
}

func StringStartsWith(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(String)
    if !validateString(vm, args[1], "Argument") { return nil }
    b := args[1].(String)
    return toBool(strings.HasPrefix(string(a), string(b)))
}
func StringEndsWith(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(String)
    if !validateString(vm, args[1], "Argument") { return nil }
    b := args[1].(String)
    return toBool(strings.HasSuffix(string(a), string(b)))
}
func StringContainsSubstring(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(String)
    if !validateString(vm, args[1], "Argument") { return nil }
    b := args[1].(String)
    return toBool(strings.Contains(string(a), string(b)))
}
// func StringPartition(vm *TenoriteVM, args []Receiver) Receiver {
//     a := args[0].(String)
//     if !validateString(vm, args[1], "Argument") { return nil }
//     b := args[1].(String)
//     first, second, ok := strings.Cut(string(a), string(b))
//     if !ok {
//         args[0] = Pair { String(first), NONE }
//         return PrimitiveSuccess
//     }
//     args[0] = Pair { String(first), String(second) }
//     return PrimitiveSuccess
// }
func StringIndexOf1(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(String)
    if !validateString(vm, args[1], "Argument") { return nil }
    b := args[1].(String)
    return Number(strings.Index(string(a), string(b)))
}
func StringIndexOf2(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(String)
    if !validateString(vm, args[1], "Argument") { return nil }
    if !validateNumber(vm, args[2], "Argument") { return nil }
    b := args[1].(String)
    start := int(args[2].(Number))
    if start >= len(a) {
        vm.Error = fmt.Errorf("Index out of bounds")
        return nil
    }
    index := strings.Index(string(a[start:]), string(b))
    if index != -1 {
        return Number(index+start)
    } else {
        return Number(-1)
    }
}
func StringRepeat(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(String)
    if !validateNumber(vm, args[1], "Argument") { return nil }
    times := args[1].(Number)
    return String(strings.Repeat(string(a), int(times)))
}
func StringSplit(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(String)
    if !validateString(vm, args[1], "Argument") { return nil }
    b := args[1].(String)

    strs := strings.Split(string(a), string(b))
    result := List { make([]Receiver, len(strs)) }
    for i, _ := range strs {
        result.List[i] = String(strs[i])
    }
    return result
}
func StringConcatString(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(String)
    if !validateString(vm, args[1], "Argument") { return nil }
    b := args[1].(String)
    return String(string(a)+string(b))
}

func StringSlice(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Argument") { return nil }
    str := args[0].(String)
    runes := []rune(str)
    start := int(args[1].(Number))
    return String(runes[start:])
}
func StringSliceEnd(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Argument") { return nil }
    if !validateNumber(vm, args[2], "Argument") { return nil }
    str := args[0].(String)
    runes := []rune(str)
    start := int(args[1].(Number))
    end := int(args[2].(Number))
    if end < 0 {
        end = len(runes)+end
    }
    return String(runes[start:end])
}

func StringString(vm *TenoriteVM, args []Receiver) Receiver {
    return args[0] }
func StringFormat(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateString(vm, args[1], "Format") { return nil }
    self := args[0].(String)
    format := args[1].(String)
    fill, right, w, rest := formatingSpec(string(format))
    if rest == "r" {
        return String(padStr(fill, right, w, fmt.Sprintf("%q", args[0])))
    }
    return String(padStr(fill, right, w, string(self)))
}

// ============ Function ============

func FunctionArity(vm *TenoriteVM, args []Receiver) Receiver {
    sub := args[0].(*Closure)
    return Number(sub.CodeObj.Arity)
}

func FunctionCall(vm *TenoriteVM, args []Receiver) Receiver {
    if p, isPrim := args[0].(Primitive); isPrim { return p.Call(vm, args) }

    sub := args[0].(*Closure)
    if len(args) < int(sub.CodeObj.Arity) {
        vm.Error = fmt.Errorf("Function expects %d arguments.", sub.CodeObj.Arity)
    }

    result, err := RunClosure(vm, sub, args)
    if err != nil {
        vm.Error = err
        return nil
    }

    return result
}

func FunctionCallWithValues(vm *TenoriteVM, args []Receiver) Receiver {

    if !validateList(vm, args[1], "Argument") { return nil }
    arglist := args[1].(List)

    if p, isPrim := args[0].(Primitive); isPrim { return p.Call(vm, append([]Receiver{p}, arglist.List...)) }

    sub := args[0].(*Closure)
    if len(arglist.List) < int(sub.CodeObj.Arity) {
        vm.Error = fmt.Errorf("Function expects %d arguments.", sub.CodeObj.Arity)
    }

    result, err := RunClosure(vm, sub, append([]Receiver{sub}, arglist.List...))
    if err != nil {
        vm.Error = err
        return nil
    }

    return result
}

// ============ List ============

func ListNewFill(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Size") { return nil }
    size := int(args[1].(Number))
    fill := args[2]
    l := make([]Receiver, size)
    for i := 0; i < size; i++ {
        l[i] = fill
    }
    return List { l }
}


func ListMakeTable(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(List)
    if !validateList(vm, args[1], "Right side") { return nil }
    b := args[1].(List)
    if len(a.List) != len(b.List) {
        vm.Error = fmt.Errorf("Keys and Values must conform.")
        return nil
    }
    return Table { b.List, a.List }
}

func ListConcat(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(List)
    if !validateList(vm, args[1], "Right side") { return nil }
    b := args[1].(List)
    return List { append(a.List, b.List...) }
}

func ListAppend(vm *TenoriteVM, args []Receiver) Receiver {
    a := args[0].(List)
    b := args[1]
    return List { append(a.List, b) }
}

func ListLen(vm *TenoriteVM, args []Receiver) Receiver {
    list := args[0].(List)
    return Number(len(list.List))
}

func ListAt_(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Argument") { return nil }
    list := args[0].(List)
    index := int(args[1].(Number))
    return list.List[index]
}

func ListCompress(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateList(vm, args[1], "Argument") { return nil }
    list := args[0].(List)
    mask := args[1].(List)
    if len(list.List) != len(mask.List) {
        vm.Error = fmt.Errorf("Differing sizes")
        return nil
    }
    var result []Receiver
    for i, r := range list.List {
        if !isFalsey(mask.List[i]) {
            result = append(result, r)
        }
    }
    return List { result }
}

func ListSliceEnd(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Argument") { return nil }
    if !validateNumber(vm, args[2], "Argument") { return nil }
    list := args[0].(List)
    start := int(args[1].(Number))
    end := int(args[2].(Number))
    if end < 0 {
        end = len(list.List)+end
    }
    return List { list.List[start:end] }
}

func ListAll(vm *TenoriteVM, args []Receiver) Receiver {
    list := args[0].(List)
    for _, r := range list.List {
        if isFalsey(r) { return FALSE }
    }
    return TRUE
}

func ListAny(vm *TenoriteVM, args []Receiver) Receiver {
    list := args[0].(List)
    for _, r := range list.List {
        if !isFalsey(r) { return TRUE }
    }
    return FALSE
}

func ListTakeNumber(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Argument") { return nil }
    list := args[0].(List)
    end := int(args[1].(Number))
    if end < 0 {
        end = len(list.List)+end
        return List { list.List[end:len(list.List)] }
    }
    return List { list.List[0:end] }
}
func ListDropNumber(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateNumber(vm, args[1], "Argument") { return nil }
    list := args[0].(List)
    start := int(args[1].(Number))
    if start < 0 {
        return List { list.List[0:len(list.List)+start] }
    }
    return List { list.List[start:len(list.List)] }
}
func ListGroup(vm *TenoriteVM, args []Receiver) Receiver {
    if !validateList(vm, args[1], "Argument") { return nil }
    list := args[0].(List)
    groups := args[1].(List)
    hash := make(map[Receiver]List)
    var keys []Receiver
    var values []Receiver
    for i, key := range groups.List {
        bucket, ok := hash[key]
        bucket.List = append(bucket.List, list.List[i])
        hash[key] = bucket
        if !ok { keys = append(keys, key) }
    }
    values = make([]Receiver, len(keys))
    for i, key := range keys {
        values[i] = hash[key]
    }
    return Table { keys, values }
}

// ============ Table ============

func TableLen(vm *TenoriteVM, args []Receiver) Receiver {
    table := args[0].(Table)
    return Number(len(table.Keys))
}
func TableKeys(vm *TenoriteVM, args []Receiver) Receiver {
    table := args[0].(Table)
    return List { table.Keys }
}
func TableValues(vm *TenoriteVM, args []Receiver) Receiver {
    table := args[0].(Table)
    return List { table.Values }
}
func TableAt_(vm *TenoriteVM, args []Receiver) Receiver {
    table := args[0].(Table)
    key := args[1]
    for i, ikey := range table.Keys {
        if sameObj(key, ikey) {
            return table.Values[i]
        }
    }
    return NONE
}

// ============ Pair ============

func PairFirst(vm *TenoriteVM, args []Receiver) Receiver {
    return args[0].(Pair).First
}
func PairSecond(vm *TenoriteVM, args []Receiver) Receiver {
    return args[0].(Pair).Second
}

// ============ Range ============

func RangeFrom(vm *TenoriteVM, args []Receiver) Receiver {
    return args[0].(Range).From }
func RangeTo(vm *TenoriteVM, args []Receiver) Receiver {
    return args[0].(Range).To }
func RangeMin(vm *TenoriteVM, args []Receiver) Receiver {
    r := args[0].(Range)
    if r.From < r.To { return r.From }
    return r.To
}
func RangeMax(vm *TenoriteVM, args []Receiver) Receiver {
    r := args[0].(Range)
    if r.From < r.To { return r.To }
    return r.From
}
func RangeLen(vm *TenoriteVM, args []Receiver) Receiver {
    r := args[0].(Range)
    if r.From < r.To { return r.To - r.From }
    return r.From - r.To
}
func RangeList(vm *TenoriteVM, args []Receiver) Receiver {
    r := args[0].(Range)
    var size int
    if r.From < r.To {
        size = int(r.To - r.From)
        slice := make([]Receiver, int(size+1))
        for i := 0; i < size+1; i++ {
            slice[i] = Number(i)+r.From
        }
        return List { slice }
    } else {
        size = int(r.From - r.To)
        slice := make([]Receiver, int(size+1))
        for i := 0; i < size+1; i++ {
            slice[i] = r.From-Number(i)
        }
        return List { slice }
    }
    
}
func RangeNext(vm *TenoriteVM, args []Receiver) Receiver {
    r := args[0].(Range)
    if isFalsey(args[1]) { return Number(r.From) }
    if !validateNumber(vm, args[1], "Argument") { return nil }
    index := args[1].(Number)
    if r.From < r.To {
        if index+1 > r.To { return NONE }
        return Number(index+1)
    }
    if index-1 < r.To { return NONE }
    return Number(index-1)
}

// ============ System ============

func SystemAssert(vm *TenoriteVM, args []Receiver) Receiver {
    if isFalsey(args[1]) {
        vm.Error = fmt.Errorf("Assertion Failed")
        return nil
    }
    return NONE
}

func SystemWriteString(vm *TenoriteVM, args []Receiver) Receiver {
    fmt.Printf("%v", args[1])
    return NONE
}

func SystemPanic(vm *TenoriteVM, args []Receiver) Receiver {
    vm.Error = fmt.Errorf("%v", args[1])
    return nil
}

// ============ Reflect ============

func ReflectListMethods(vm *TenoriteVM, args []Receiver) Receiver {
    ns := args[1].(*Namespace)
    var methods []Receiver
    for sym, _ := range ns.Table {
        methods = append(methods, sym)
    }
    return List { methods }
}

func ReflectNotRespondsTo(vm *TenoriteVM, args []Receiver) Receiver {
    obj := args[1]
    meth := args[2].(Symbol)
    if obj.GetMethod(meth) != nil {
        return FALSE
    }
    return TRUE
}

func PreInitializeCore(vm *TenoriteVM) {
    coreMod := vm.NewModule("")
    vm.TopModule = coreMod

    coreMod.Add(vm.Symbol("True"), TRUE)
    coreMod.Add(vm.Symbol("False"), FALSE)
    coreMod.Add(vm.Symbol("None"), NONE)

    coreMod.Add(vm.Symbol("Object"), ObjectNs)
    coreMod.Add(vm.Symbol("Bool"), BoolNs)
    coreMod.Add(vm.Symbol("Number"), NumberNs)
    coreMod.Add(vm.Symbol("String"), StringNs)
    coreMod.Add(vm.Symbol("Function"), FunctionNs)
    coreMod.Add(vm.Symbol("List"), ListNs)
    coreMod.Add(vm.Symbol("Table"), TableNs)
    coreMod.Add(vm.Symbol("Range"), RangeNs)

    coreMod.Add(vm.Symbol("Eq"), EqNs)
    coreMod.Add(vm.Symbol("Ord"), OrdNs)

    coreMod.Add(vm.Symbol("formatPadding"), Primitive { FormatPadding })

    ObjectNs.Set(vm.Symbol("==="), Primitive { ObjSame })
    ObjectNs.Set(vm.Symbol("!=="), Primitive { ObjNotSame })
    ObjectNs.Set(vm.Symbol("=>"), Primitive { ObjPair })
    ObjectNs.Set(vm.Symbol("string"), Primitive { ObjString })

    NamespaceNs.Set(vm.Symbol("new:"), Primitive { NamespaceMake })
    
    BoolNs.Set(vm.Symbol("=="), Primitive { ObjSame })
    BoolNs.Set(vm.Symbol("!="), Primitive { ObjNotSame })
    BoolNs.Set(vm.Symbol("not"), Primitive { BoolNot })
    BoolNs.Set(vm.Symbol("string"), Primitive { BoolString })

    NumberNs.Set(vm.Symbol("+"), Primitive { NumAdd })
    NumberNs.Set(vm.Symbol("-"), Primitive { NumSubtract })
    NumberNs.Set(vm.Symbol("*"), Primitive { NumMultiply })
    NumberNs.Set(vm.Symbol("/"), Primitive { NumDivide })
    NumberNs.Set(vm.Symbol("%"), Primitive { NumModulo })
    NumberNs.Set(vm.Symbol("**"), Primitive { NumPower })
    NumberNs.Set(vm.Symbol(">"), Primitive { NumGt })
    NumberNs.Set(vm.Symbol("<"), Primitive { NumLt })
    NumberNs.Set(vm.Symbol(">="), Primitive { NumGe })
    NumberNs.Set(vm.Symbol("<="), Primitive { NumLe })
    NumberNs.Set(vm.Symbol("=="), Primitive { ObjSame })
    NumberNs.Set(vm.Symbol("!="), Primitive { ObjNotSame })
    NumberNs.Set(vm.Symbol(";"), Primitive { NumRange })
    NumberNs.Set(vm.Symbol(">>"), Primitive { NumMax })
    NumberNs.Set(vm.Symbol("<<"), Primitive { NumMin })
    NumberNs.Set(vm.Symbol("string"), Primitive { NumString })

    StringNs.Set(vm.Symbol(">"), Primitive { StringGt })
    StringNs.Set(vm.Symbol("<"), Primitive { StringLt })
    StringNs.Set(vm.Symbol(">="), Primitive { StringGe })
    StringNs.Set(vm.Symbol("<="), Primitive { StringLe })
    StringNs.Set(vm.Symbol("=="), Primitive { ObjSame })
    StringNs.Set(vm.Symbol("!="), Primitive { ObjNotSame })
    StringNs.Set(vm.Symbol("len"), Primitive { StringLen })
    StringNs.Set(vm.Symbol("upper"), Primitive { StringUpper })
    StringNs.Set(vm.Symbol("lower"), Primitive { StringLower })
    StringNs.Set(vm.Symbol("trim"), Primitive { StringTrim })
    StringNs.Set(vm.Symbol("explode"), Primitive { StringExplode })
    StringNs.Set(vm.Symbol("trimLeft"), Primitive { StringTrimLeft })
    StringNs.Set(vm.Symbol("trimRight"), Primitive { StringTrimRight })
    StringNs.Set(vm.Symbol("at_:"), Primitive { StringAt_ })
    StringNs.Set(vm.Symbol("startsWith:"), Primitive { StringStartsWith })
    StringNs.Set(vm.Symbol("endsWith:"), Primitive { StringEndsWith })
    StringNs.Set(vm.Symbol("containsString:"), Primitive { StringContainsSubstring })
    StringNs.Set(vm.Symbol("indexOfString:"), Primitive { StringIndexOf1 })
    StringNs.Set(vm.Symbol("indexOfString:start:"), Primitive { StringIndexOf2 })
    StringNs.Set(vm.Symbol("repeat:"), Primitive { StringRepeat })
    StringNs.Set(vm.Symbol("split:"), Primitive { StringSplit })
    StringNs.Set(vm.Symbol("concatString:"), Primitive { StringConcatString })
    StringNs.Set(vm.Symbol("slice:"), Primitive { StringSlice })
    StringNs.Set(vm.Symbol("slice:end:"), Primitive { StringSliceEnd })
    StringNs.Set(vm.Symbol("%%"), Primitive { StringFormat })
    StringNs.Set(vm.Symbol("string"), Primitive { StringString })

    StringNs.Set(vm.Symbol("findRegex:"), Primitive { StringFindRegex })
    StringNs.Set(vm.Symbol("findRegex:start:"), Primitive { StringFindRegexStart })

    FunctionNs.Set(vm.Symbol("arity"), Primitive { FunctionArity })
    FunctionNs.Set(vm.Symbol("call"), Primitive { FunctionCall })
    FunctionNs.Set(vm.Symbol("value:"), Primitive { FunctionCall })
    FunctionNs.Set(vm.Symbol("value:value:"), Primitive { FunctionCall })
    FunctionNs.Set(vm.Symbol("value:value:value:"), Primitive { FunctionCall })
    FunctionNs.Set(vm.Symbol("value:value:value:value:"), Primitive { FunctionCall })
    FunctionNs.Set(vm.Symbol("value:value:value:value:value:"), Primitive { FunctionCall })
    FunctionNs.Set(vm.Symbol("value:value:value:value:value:value:"), Primitive { FunctionCall })
    FunctionNs.Set(vm.Symbol("callWithValues:"), Primitive { FunctionCallWithValues })

    ListNs.Static = NewNamespace("")
    ListNs.Static.Set(vm.Symbol("new:fill:"), Primitive { ListNewFill })
    ListNs.Set(vm.Symbol("!"), Primitive { ListMakeTable })
    ListNs.Set(vm.Symbol("<>"), Primitive { ListConcat })
    ListNs.Set(vm.Symbol("++"), Primitive { ListAppend })
    ListNs.Set(vm.Symbol("len"), Primitive { ListLen })
    ListNs.Set(vm.Symbol("all"), Primitive { ListAll })
    ListNs.Set(vm.Symbol("any"), Primitive { ListAny })
    ListNs.Set(vm.Symbol("takeNumber:"), Primitive { ListTakeNumber })
    ListNs.Set(vm.Symbol("dropNumber:"), Primitive { ListDropNumber })
    ListNs.Set(vm.Symbol("at_:"), Primitive { ListAt_ })
    ListNs.Set(vm.Symbol("compress:"), Primitive { ListCompress })
    ListNs.Set(vm.Symbol("slice:end:"), Primitive { ListSliceEnd })
    ListNs.Set(vm.Symbol("groupList:"), Primitive { ListGroup })

    TableNs.Set(vm.Symbol("len"), Primitive { TableLen })
    TableNs.Set(vm.Symbol("keys"), Primitive { TableKeys })
    TableNs.Set(vm.Symbol("values"), Primitive { TableValues })
    TableNs.Set(vm.Symbol("at_:"), Primitive { TableAt_ })

    PairNs.Set(vm.Symbol("first"), Primitive { PairFirst })
    PairNs.Set(vm.Symbol("second"), Primitive { PairSecond })

    RangeNs.Set(vm.Symbol("from"), Primitive { RangeFrom })
    RangeNs.Set(vm.Symbol("to"), Primitive { RangeTo })
    RangeNs.Set(vm.Symbol("min"), Primitive { RangeMin })
    RangeNs.Set(vm.Symbol("max"), Primitive { RangeMax })
    RangeNs.Set(vm.Symbol("len"), Primitive { RangeLen })
    RangeNs.Set(vm.Symbol("list"), Primitive { RangeList })
    RangeNs.Set(vm.Symbol("next:"), Primitive { RangeNext })

    RegexNs.Static = RegexNs
    coreMod.Add(vm.Symbol("RegexResults"), RegexResultsNs)
    coreMod.Add(vm.Symbol("Regex"), RegexNs)
    RegexNs.Set(vm.Symbol("new:"), Primitive { RegexNew })

    SystemNs := NewNamespace("System")
    SystemNs.Static = SystemNs
    coreMod.Add(vm.Symbol("System"), SystemNs)

    SystemNs.Set(vm.Symbol("assert:"), Primitive { SystemAssert })
    SystemNs.Set(vm.Symbol("panic:"), Primitive { SystemPanic })
    SystemNs.Set(vm.Symbol("writeString:"), Primitive { SystemWriteString })


    ReflectNs := NewNamespace("Reflect")
    ReflectNs.Static = ReflectNs
    coreMod.Add(vm.Symbol("Reflect"), ReflectNs)
    ReflectNs.Set(vm.Symbol("listMethods:"), Primitive { ReflectListMethods })
    ReflectNs.Set(vm.Symbol("notResponds:to:"), Primitive { ReflectNotRespondsTo })
}

func InitializeCore(vm *TenoriteVM) {
    // coreMod := vm.TopModule
}