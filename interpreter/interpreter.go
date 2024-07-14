package interpreter

import (
    "fmt"
    "runtime"
    "reflect"
)

func Run(vm *TenoriteVM, callable Receiver, args []Receiver) (Receiver, error) {
    switch sub := callable.(type) {
    case *Closure:
        return RunClosure(vm, sub, args)
    case Primitive:
        result := sub.Call(vm, args)
        if result == nil {
            return nil, fmt.Errorf("%v %s",
                runtime.FuncForPC(reflect.ValueOf(sub.Call).Pointer()).Name(),
                vm.Error)
        }
        return result, nil
    }
    return nil, fmt.Errorf("Not a callable")
}

func getLine(ip int, lines []int) (line int) {
    for i := 1; i < len(lines); i++ {
        if ip > lines[i] { continue }
        line = i
        break
    }
    return
}

func RunClosure(vm *TenoriteVM, sub *Closure, args []Receiver) (Receiver, error) {
    code := sub.CodeObj.Code
    ip := 0

    task := Task {}

    locals := make([]Receiver, int(sub.CodeObj.LocalSize)+len(args))
    for i, arg := range args {
        locals[i] = arg
    }

    task.Push(locals[0])

    for {
        debugIp := ip
        debugName := sub.CodeObj.Name
        debugLines := sub.CodeObj.Lines

        op := code[ip]
        switch op {
        case OP_NOP:
            ip++
        case OP_CONST:
            task.Push(sub.CodeObj.Consts[code[ip+1]])
            ip+=2
        case OP_SYM:
            task.Push(Symbol(code[ip+1]))
            ip+=2
        case OP_CLOSURE:
            task.Push(NONE)
            codeObj := sub.CodeObj.Consts[code[ip+1]].(*CodeObj)
            closure := &Closure{ codeObj, make([]*Upvalue, codeObj.UpvalueCount) }
            ip++
            for i := uint16(0); i < codeObj.UpvalueCount; i++ {
                ip++
                isLocal := code[ip]
                ip++
                index := code[ip]
                if isLocal != 0 {
                    closure.Upvalues[i] = captureUpvalue(&task, locals, index)
                } else {
                    closure.Upvalues[i] = sub.Upvalues[index]
                }
            }
            task.Stack[len(task.Stack)-1] = closure
            // task.Push(closure)
            ip++
        case OP_STORE_MODULE:
            name := Symbol(code[ip+1])
            value := task.Stack[len(task.Stack)-1]
            loc, ok := vm.TopModule.Table[name]
            if !ok {
                vm.TopModule.Add(name, value)
            } else {
                vm.TopModule.Variables[loc] = value
            }
            ip+=2
        case OP_LOAD_MODULE:
            name := Symbol(code[ip+1])
            loc, ok := vm.TopModule.Table[name]
            if !ok {
                task.Error = fmt.Errorf("Undefined Name #%s", vm.SymbolStore[name])
                task.Panic(ip, debugLines)
            }
            task.Push(vm.TopModule.Variables[loc])
            ip+=2
        case OP_STORE_LOCAL:
            at := code[ip+1]
            // locals[at] = task.Stack[len(task.Stack)-1]
            locals[at] = task.Pop()
            ip+=2
        case OP_LOAD_LOCAL:
            at := code[ip+1]
            task.Push(locals[at])
            ip+=2
        case OP_LOAD_UPVALUE:
            at := code[ip+1]
            upvalue := sub.Upvalues[at]
            task.Push(*upvalue.Value)
            ip+=2
        case OP_STORE_UPVALUE:
            at := code[ip+1]
            upvalue := sub.Upvalues[at]
            *upvalue.Value = task.Pop()
            ip+=2
        case OP_LOAD_FIELD:
            name := Symbol(code[ip+1])
            obj, ok := locals[0].(Object)
            if !ok {
                task.Error = fmt.Errorf("Invalid field `%s´ access", vm.SymbolStore[name])
                task.Panic(ip, debugLines)
            }
            result := obj.Table[name]
            if result == nil {
                task.Error = fmt.Errorf("Invalid field `%s´ access", vm.SymbolStore[name])
                task.Panic(ip, debugLines)
            }
            task.Push(result)
            ip+=2
        case OP_STORE_FIELD:
            name := Symbol(code[ip+1])
            obj, ok := locals[0].(Object)
            if !ok {
                task.Error = fmt.Errorf("Invalid field access")
                task.Panic(ip, debugLines)
            }
            obj.Table[name] = task.Stack[len(task.Stack)-1]
            ip+=2
        /*case OP_OPERATOR:
            name := Symbol(code[ip+1])
            op, ok := vm.Operators[name]
            if !ok {
                task.Error = fmt.Errorf("Undefined Operator #%s", vm.SymbolStore[name])
                task.Panic(ip, debugLines)
            }
            b := task.Pop()
            a := task.Pop()

            if !op(vm, a, b) {
                vm.Task.Panic(ip)
            }
            ip+=2*/
        case OP_LOOP:
            bool := task.Pop()
            jump_backwards := code[ip+1]
            if _, isFalse := bool.(False); isFalse {
                ip+=2
            } else {
                ip-=int(jump_backwards)+2
            }
        case OP_JUMP_FALSE:
            bool := task.Pop()
            jump_forward := code[ip+1]
            if _, isFalse := bool.(False); isFalse {
                ip+=int(jump_forward)
            } else {
                ip+=2
            }
        case OP_CALL_R:
            nargs := int(code[ip+1])+1
            fp := len(task.Stack)-(nargs+1)
            ip+=2

            sym := task.Pop().(Symbol)
            msg := Message { sym, make([]int, nargs) }

            for i := 0; i < nargs; i++ {
                msg.Ranks[i] = int(code[ip])
                ip++
            }

            result, err := Call(vm, msg, task.Stack[fp:])
            if err != nil {
                task.Error = err
                task.Panic(ip, debugLines)
            }

            task.Stack = task.Stack[:fp]
            task.Push(result)

        case OP_CALL_0R1:
            fp := len(task.Stack)-3
            ip+=1

            sym := task.Pop().(Symbol)
            msg := Message { sym, []int{0, 1} }
            
            result, err := Call(vm, msg, task.Stack[fp:])
            if err != nil {
                task.Error = err
                task.Panic(ip, debugLines)
            }

            task.Stack = task.Stack[:fp]
            task.Push(result)
        case OP_CALL:
            nargs := int(code[ip+1])+1
            fp := len(task.Stack)-(nargs+1)
            ip+=2

            sym := task.Pop().(Symbol)
            msg := Message { sym, make([]int, nargs) }

            result, err := Call(vm, msg, task.Stack[fp:])
            if err != nil {
                task.Error = err
                task.Panic(ip, debugLines)
            }

            task.Stack = task.Stack[:fp]
            task.Push(result)
        case OP_POP:
            task.Pop()
            ip+=1
        case OP_PRINT:
            r := task.Pop()
            if s, isString := r.(String); isString {
                fmt.Printf("%s\n", s)
            }
            ip+=1
        case OP_CLOSE_UPVALUE:
            at := code[ip+1]
            closeUpvalue(&task, locals, at)
            ip+=2
        case OP_RETURN:
            result := task.Pop()
            return result, nil
        case OP_TYPE:
            ns := task.Pop()
            obj := task.Pop()
            task.Push(toBool(obj.Type(ns)))
            ip+=1
        case OP_MAKE_LIST:
            ip++
            n := code[ip]
            list := make([]Receiver, n)
            for i := uint16(0); i < n; i++ {
                val := task.Pop()
                list[n-i-1] = val
            }
            task.Push(List { list })
            ip++
        case OP_MAKE_NS:
            name := task.Pop().(String)
            ns := NewNamespace(string(name))
            task.Push(ns)
            ip+=1
        case OP_MAKE_METHOD:
            subroutine := task.Pop().(*Closure)
            symbol := Symbol(code[ip+1])
            obj := task.Pop()
            if ns, ok := obj.(*Namespace); ok {
                ns.Table[symbol] = subroutine
            } else {
                task.Error = fmt.Errorf("Object not a namespace %v", obj)
                task.Panic(ip, debugLines)
            }
            task.Push(obj)
            ip+=2
        case OP_MAKE_STATIC:
            subroutine := task.Pop().(*Closure)
            symbol := Symbol(code[ip+1])
            obj := task.Pop()
            if ns, ok := obj.(*Namespace); ok {
                if ns.Static == nil {
                    ns.Static = NewNamespace("")
                }
                ns.Static.Table[symbol] = subroutine
            } else {
                task.Error = fmt.Errorf("Object not a namespace %v", obj)
                task.Panic(ip, debugLines)
            }
            task.Push(obj)
            ip+=2
        /*case OP_MAKE_TABLE:
            ip++
            n := code[ip]
            table := make(map[Receiver]Receiver, n)
            for i := uint16(0); i < n; i++ {
                val := task.Pop()
                key := task.Pop()
                if !isHashable(key) {
                    task.Error = fmt.Errorf("Invalid type for table key")
                    task.Panic(ip)
                }
                table[key] = val
            }
            task.Push(Table { table, nil })
            ip++
        case OP_MAKE_CONS:
            subroutine := task.Pop().(*Closure)
            symbol := Symbol(code[ip+1])
            ns := task.Pop().Namespace()
            ns.Table[symbol] = subroutine
            task.Push(ns)
            ip+=2
        case OP_MAKE_OBJ:
            ns := task.Stack[stackStart].(*Namespace)
            task.Stack[stackStart] = Object { ns, make(map[Symbol]Receiver, 0) }
            ip+=1
        */
        case OP_RECURSIVE:
            task.Stack = task.Stack[:1]
            ip = 0
        default:
            task.Error = fmt.Errorf("Invalid Opcode %s", OPCODE_NAMES[op])
            task.Panic(ip, debugLines)
            return nil, nil
        }

        if vm.StackTrace {
            fmt.Printf("%4d %-15s %-15s [ ", getLine(debugIp, debugLines), debugName, OPCODE_NAMES[code[debugIp]])
            for _, v := range task.Stack {
                if v == nil {
                    fmt.Printf("<nil> ")
                    continue
                }
                fmt.Printf("%v ", toDebugString(vm, v))
            }
            fmt.Printf("]\n")
        }
    }
}


func captureUpvalue(task *Task, locals []Receiver, slot uint16) *Upvalue {
    var prevUpvalue *Upvalue
    var upvalue *Upvalue = task.OpenUpvalues
    for upvalue != nil && upvalue.Slot > slot {
        prevUpvalue = upvalue
        upvalue = upvalue.Next
    }

    if upvalue != nil && upvalue.Slot > slot {
        return upvalue
    }

    newUpvalue := &Upvalue {
        // Value: &task.Stack[slot],
        Value: &locals[slot],
        Slot: slot,
        Next: upvalue,
    }

    if prevUpvalue == nil {
        task.OpenUpvalues = newUpvalue
    } else {
        prevUpvalue.Next = newUpvalue
    }

    return newUpvalue
}

func closeUpvalue(task *Task, locals []Receiver, slot uint16) {
    for task.OpenUpvalues != nil && task.OpenUpvalues.Slot >= slot {
        upvalue := task.OpenUpvalues
        upvalue.Closed = *upvalue.Value
        upvalue.Value = &upvalue.Closed
        task.OpenUpvalues = upvalue.Next
    }
}

type Task struct {
    Stack         []Receiver
    Error         error
    OpenUpvalues  *Upvalue
}

func (task *Task) Push(value Receiver)  {
    task.Stack = append(task.Stack, value)
}

func (task *Task) Pop() (value Receiver) {
    value = task.Stack[len(task.Stack)-1]
    task.Stack = task.Stack[:len(task.Stack)-1]
    return
}

func (task *Task) Panic(ip int, lines []int) {
    line := getLine(ip, lines)
    panic(fmt.Sprintf("line %d: %s", line, task.Error.Error()))
}