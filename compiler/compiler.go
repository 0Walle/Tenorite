package compiler

import (
    // "io"
    "strconv"
    "fmt"
    // "strings"
    "regexp"
    _ "embed"
    // "os"
    "0Walle/Tenorite/interpreter"
    "0Walle/Tenorite/parser"
    "0Walle/Tenorite/token"
)

type PrintReporter struct {}

func (r PrintReporter) Report(line int, where string, message string)  {
    fmt.Printf("Line %d: at `%s´: %s\n", line, where, message)
}

type CompilerState struct {
    VM     *interpreter.TenoriteVM
    Frames  []StackFrame
    Frame   *StackFrame
    Subs    []*interpreter.CodeObj
}

type Upvalue struct {
    Index    uint16
    IsLocal  bool
}

type Local struct {
    Slot        uint16
    IsCaptured  bool
}

type StackFrame struct {
    Environment   map[string]Local
    Upvalues      []Upvalue
    Sub           *interpreter.CodeObj
    Last          *StackFrame
    Lines         []int
}

func (fr *StackFrame) Write(code ...uint16) int {
    fr.Sub.Code = append(fr.Sub.Code, code...)
    return len(fr.Sub.Code)-1
}

func (fr *StackFrame) WriteClosure(loc uint16, upvalues []Upvalue) {
    fr.Write(interpreter.OP_CLOSURE, loc)
    for _, upvalue := range upvalues {
        if upvalue.IsLocal {
            fr.Write(1, upvalue.Index)
        } else {
            fr.Write(0, upvalue.Index)
        }
    }
}

func (comp *CompilerState) PushFrame(recv string, params []string, subName string) {
    comp.Subs = append(comp.Subs, &interpreter.CodeObj {})
    sub := comp.Subs[len(comp.Subs)-1]

    sub.Arity = uint16(len(params))
    sub.Name = subName

    comp.Frames = append(comp.Frames, StackFrame {
        Environment: make(map[string]Local, 0),
        Last: comp.Frame,
        Sub: sub,
    })
    comp.Frame = &comp.Frames[len(comp.Frames)-1]

    if recv != "" {
        comp.Frame.Environment[recv] = Local { 0, false }
    }

    for i, name := range params {
        comp.Frame.Environment[name] = Local { uint16(i)+1, false }
    }
}

func (comp *CompilerState) PopFrame() (*interpreter.CodeObj, []Upvalue) {
    sub := comp.Frame.Sub
    sub.CoVarnames = make([]string, sub.LocalSize+sub.Arity+1)
    for sym, loc := range comp.Frame.Environment {
        sub.CoVarnames[loc.Slot] = sym
    }

    sub.Lines = comp.Frame.Lines

    upvalues := comp.Frame.Upvalues
    comp.Frames = comp.Frames[:len(comp.Frames)-1]
    if len(comp.Frames) != 0 {
        comp.Frame = &comp.Frames[len(comp.Frames)-1]
    }
    return sub, upvalues
}

func (comp *CompilerState) PushConst(c interpreter.Receiver) uint16 {
    for loc, val := range comp.Frame.Sub.Consts {
        if c == val {
            return uint16(loc)
        }
    }

    loc := uint16(len(comp.Frame.Sub.Consts))
    comp.Frame.Sub.Consts = append(comp.Frame.Sub.Consts, c)
    return loc
}

func (comp *CompilerState) AddLine(line int) {
    if line >= len(comp.Frame.Lines) {
        new := make([]int, 0, line+1)
        new = append(new, comp.Frame.Lines...)
        comp.Frame.Lines = new[:line+1]
    }

    // fmt.Printf("Add line %d ip %d\n", line, len(comp.Frame.Sub.Opcodes))
    // fmt.Printf("%#v\n", comp.Frame.Lines)

    comp.Frame.Lines[line] = len(comp.Frame.Sub.Code)
}


//go:embed core.tenor
var coreInc string

func CompileCore(ctx *interpreter.TenoriteVM) {
    interpreter.PreInitializeCore(ctx)

    coreInc += "\n"

    scanner := token.NewScanner(coreInc, PrintReporter{})
    tokens, hasError := scanner.Scan()
    if hasError {
        panic("Scanner error")
    }

    parser := parser.NewParser(tokens)
    unit, err := parser.ParseUnit()
    if err != nil { panic(err) }

    comp := CompilerState { VM: ctx }

    comp.PushFrame("", nil, "__core__")

    err = comp.CompileModule(unit)
    if err != nil { panic(err) }

    // for i, _ := range comp.Subs {
    //     vm.PrintSub(ctx, comp.Subs[i])
    // }

    sub, _ := comp.PopFrame()
    closure := &interpreter.Closure{ sub, nil }
    interpreter.RunClosure(ctx, closure, []interpreter.Receiver{interpreter.NONE})
}

func Compile(source string, unitName string) {


    ctx := interpreter.MakeVM()
    CompileCore(&ctx)
    

    scanner := token.NewScanner(source, PrintReporter{})
    tokens, hasError := scanner.Scan()
    if hasError {
        panic("Scanner error")
    }

    // for _, tk := range tokens {
    //     fmt.Printf("%d`%s´ ", tk.Kind, tk.Lexeme)
    // }
    // fmt.Printf("\n")

    parser := parser.NewParser(tokens)
    unit, err := parser.ParseUnit()
    if err != nil {
        panic(err)
    }

    // fmt.Printf("\n")
    // unit.Print(0)
    // fmt.Printf("\n\n")

    comp := CompilerState {
        VM: &ctx,
    }

    main := ctx.NewModule("__main__")
    ctx.TopModule = main

    coreMod := ctx.Modules[""]

    for name, loc := range coreMod.Table {
        
        // fmt.Printf("Imported %s\n", ctx.SymbolStore[name])
        ctx.TopModule.Add(name, coreMod.Variables[loc])
    }

    comp.PushFrame("", nil, unitName)

    err = comp.CompileModule(unit)
    if err != nil {
        panic(err)
    }

    sub, _ := comp.PopFrame()

    // for i, _ := range comp.Subs {
    //     vm.PrintSub(&ctx, comp.Subs[i])
    // }

    // fmt.Printf("\n")

    // ctx.StackTrace = true

    closure := &interpreter.Closure{ sub, nil }
    result, err := interpreter.RunClosure(&ctx, closure, []interpreter.Receiver{interpreter.NONE})
    if err != nil {
        panic(err)
    }

    string := interpreter.CallUnsafe(&ctx, result, "string", nil)
    fmt.Printf("%v\n", string)
}


func (comp *CompilerState) CompileModule(unit parser.Unit) error {
    for i, stmt := range unit.Contents {
        err := comp.CompileTopLevelStmt(stmt, i == len(unit.Contents)-1)
        if err != nil { return err }
    }
    comp.Frame.Write(interpreter.OP_RETURN, interpreter.OP_END)
    return nil
}

func (comp *CompilerState) CompileTopLevelStmt(stmt parser.Stmt, isLast bool) error {
    switch stmt := stmt.(type) {
    case parser.AssignStmt:
        if stmt.NonLocal {
            return fmt.Errorf("Invalid nonlocal assignment in top level of module")
        }

        name := stmt.Name.Value.Lexeme

        sym := comp.VM.Symbol(name)

        err := comp.CompileExpr(stmt.Value)
        if err != nil { return err }

        comp.VM.TopModule.Reserve(sym)

        comp.Frame.Write(interpreter.OP_STORE_MODULE, uint16(sym))
        if !isLast { comp.Frame.Write(interpreter.OP_POP) }
        return nil
    case parser.MethStmt:
        ns := stmt.Namespace.Value.Lexeme

        symbol_, params, err := validateSelector(stmt.Params)
        if err != nil { return err }

        selector := comp.VM.Symbol(symbol_)

        recv := params[0]
        comp.PushFrame(recv, params[1:], fmt.Sprintf("%s#%s", ns, symbol_))
        
        err = comp.CompileStmtList(stmt.Body)
        if err != nil { return err }
        comp.Frame.Write(interpreter.OP_RETURN, interpreter.OP_END)
        
        sub, upvalues := comp.PopFrame()
        
        nssym := comp.VM.Symbol(ns)
        _, ok := comp.VM.TopModule.Table[nssym]
        if !ok {
            return fmt.Errorf("No such name %s in method", ns)
        }
        comp.Frame.Write(interpreter.OP_LOAD_MODULE, uint16(nssym))

        comp.Frame.WriteClosure(comp.PushConst(sub), upvalues)

        if recv == ns {
            comp.Frame.Write(interpreter.OP_MAKE_STATIC, uint16(selector))
        } else {
            comp.Frame.Write(interpreter.OP_MAKE_METHOD, uint16(selector))
        }
        
        comp.Frame.Write(interpreter.OP_POP)
    case parser.TypeStmt:
        ns := stmt.Namespace.Value
    
        nssym := comp.VM.Symbol(ns)
        
        comp.Frame.Write(interpreter.OP_CONST, comp.PushConst(interpreter.String(ns)))
        comp.Frame.Write(interpreter.OP_MAKE_NS)
        comp.VM.TopModule.Reserve(nssym)
        comp.Frame.Write(interpreter.OP_STORE_MODULE, uint16(nssym))
    case parser.LoopStmt:
        return fmt.Errorf("Invalid statement in top level of module")
    case parser.ReturnStmt:
        return fmt.Errorf("Invalid statement in top level of module")
    case parser.ExprStmt:
        err := comp.CompileExpr(stmt.X)
        if err != nil { return err }
        if !isLast {
            comp.Frame.Write(interpreter.OP_POP)
        }
        comp.AddLine(stmt.X.Line())
    }
    return nil
}

func (comp *CompilerState) CompileStmtList(list []parser.Stmt) error {
    for i, stmt := range list {
        err := comp.CompileStmt(stmt, i == len(list)-1)
        if err != nil { return err }
    }
    return nil
}

func (comp *CompilerState) CompileStmt(stmt parser.Stmt, isLast bool) error {
    switch stmt := stmt.(type) {
    case parser.AssignStmt:
        return comp.CompileAssignStmt(stmt, isLast)
    case parser.FieldAssignStmt:
        name := stmt.Name.Value.Value
        err := comp.CompileExpr(stmt.Value)
        if err != nil { return err }
        comp.Frame.Write(interpreter.OP_STORE_FIELD, uint16(comp.VM.Symbol(name)))
        if !isLast { comp.Frame.Write(interpreter.OP_POP) }
        return nil
        
    case parser.MethStmt:
        return fmt.Errorf("Invalid statement outside top level of module")
    case parser.LoopStmt:
        comp.Frame.Write(interpreter.OP_RECURSIVE)
        return nil
    case parser.ReturnStmt: return comp.CompileReturnStmt(stmt)
    case parser.ExprStmt:
        err := comp.CompileExpr(stmt.X)
        if err != nil { return err }
        if !isLast {
            if _, ok := stmt.X.(parser.StringInterpExpr); ok {
                comp.Frame.Write(interpreter.OP_PRINT)
            } else {
                comp.Frame.Write(interpreter.OP_POP)
            }
        }
        comp.AddLine(stmt.X.Line())
    }
    return nil
}

func (comp *CompilerState) CompileAssignStmt(stmt parser.AssignStmt, isLast bool) error {
    name := stmt.Name.Value.Lexeme

    comp.AddLine(stmt.Name.Value.Line)

    if !stmt.NonLocal {
        location, ok := comp.Frame.Environment[name]
        if !ok {
            location = Local {
                Slot: comp.Frame.Sub.LocalSize+comp.Frame.Sub.Arity+1,
            }
            comp.Frame.Environment[name] = location
            comp.Frame.Sub.LocalSize += 1
            // return comp.CompileExpr(stmt.Value)
        }

        err := comp.CompileExpr(stmt.Value)
        if err != nil { return err }
        comp.Frame.Write(interpreter.OP_STORE_LOCAL, location.Slot)
        // if !isLast { comp.Frame.Write(interpreter.OP_POP) }
        if isLast { comp.Frame.Write(interpreter.OP_LOAD_LOCAL, location.Slot) }
        return nil
    }

    err := comp.CompileExpr(stmt.Value)
    if err != nil { return err }

    upvalue := findUpvalue(comp, comp.Frame, name)
    if upvalue != -1 {
        comp.Frame.Write(interpreter.OP_STORE_UPVALUE, uint16(upvalue))
        return nil
    }

    sym := comp.VM.Symbol(name)
    _, ok := comp.VM.TopModule.Table[sym]
    if !ok {
        return fmt.Errorf("No such name %s in nonlocal", name)
    }
    comp.Frame.Write(interpreter.OP_STORE_MODULE, uint16(sym))
    return nil
}

func (comp *CompilerState) CompileReturnStmt(stmt parser.ReturnStmt) error {
    err := comp.CompileExpr(stmt.Cond)
    if err != nil { return err }

    label := comp.Frame.Write(interpreter.OP_JUMP_FALSE, 0)

    err = comp.CompileExpr(stmt.Return)
    if err != nil { return err }

    end := comp.Frame.Write(interpreter.OP_RETURN)

    comp.Frame.Sub.Code[label] = uint16(end-label+2)
    return nil
}

func (comp *CompilerState) CompileExpr(expr parser.Expr) error {
    switch expr := expr.(type) {
    case parser.CallExpr:
        err := comp.CompileExpr(expr.Recv)
        if err != nil { return err }

        nargs := uint16(len(expr.Args))
        
        sym := ""
        for _, kv := range expr.Args {
            sym += kv.Key.Value.Value
        }
        callsym := comp.VM.Symbol(sym)

        ranks := make([]int, len(expr.Args)+1)
        ranks[0] = expr.RRank
        s := expr.RRank
        for i, kv := range expr.Args {
            err := comp.CompileExpr(kv.Value)
            if err != nil { return err }

            ranks[i+1] = kv.Rank
            s += kv.Rank
        }
        
        comp.Frame.Write(interpreter.OP_SYM, uint16(callsym))
        if s != 0 {
            comp.Frame.Write(interpreter.OP_CALL_R, nargs)
            for i := uint16(0); i < nargs+1; i++ {
                comp.Frame.Write(uint16(ranks[i]))
            }
        } else {
            comp.Frame.Write(interpreter.OP_CALL, nargs)
        }
    case parser.BinaryExpr:
        op := expr.Op.Op.Lexeme
        sym := comp.VM.Symbol(op)
        
        err := comp.CompileExpr(expr.X)
        if err != nil { return err }
        err = comp.CompileExpr(expr.Y)
        if err != nil { return err }

        if expr.Op.Op.Kind == token.TYPE {
            comp.Frame.Write(interpreter.OP_TYPE)
        } else {
            comp.Frame.Write(interpreter.OP_SYM, uint16(sym))
            if expr.YRank == 0 && expr.XRank == 0 {
                comp.Frame.Write(interpreter.OP_CALL, 1)
            } else {
                comp.Frame.Write(interpreter.OP_CALL_R, 1)
                comp.Frame.Write(uint16(expr.XRank))
                comp.Frame.Write(uint16(expr.YRank))
            }
        }

        
    case parser.UnaryExpr:
        name := expr.Method.Value.Lexeme
        sym := comp.VM.Symbol(name)

        err := comp.CompileExpr(expr.X)
        if err != nil { return err }

        comp.Frame.Write(interpreter.OP_SYM, uint16(sym))
        if expr.XRank == 0 {
            comp.Frame.Write(interpreter.OP_CALL, 0)
        } else {
            comp.Frame.Write(interpreter.OP_CALL_R, 0, uint16(expr.XRank))
        }
    case parser.IndexExpr:
        err := comp.CompileExpr(expr.X)
        if err != nil { return err }
        err = comp.CompileExpr(expr.Y)
        if err != nil { return err }
        sym := comp.VM.Symbol("at:")
        comp.Frame.Write(interpreter.OP_SYM, uint16(sym))
        comp.Frame.Write(interpreter.OP_CALL, 1)
    case parser.ParenExpr:
        return comp.CompileExpr(expr.X)
    case parser.FunctionLiteral:
        params, err := validateParams(expr.Params)
        if err != nil { return err }
        
        comp.PushFrame("", params, fmt.Sprintf("(lambda:%d)", int(expr.Line())))

        err = comp.CompileStmtList(expr.Body)
        if err != nil { return err }

        for _, local := range comp.Frame.Environment {
            if local.IsCaptured {
                comp.Frame.Write(interpreter.OP_CLOSE_UPVALUE, local.Slot)
            }
        }

        comp.Frame.Write(interpreter.OP_RETURN, interpreter.OP_END)

        sub, upvalues := comp.PopFrame()

        comp.Frame.WriteClosure(comp.PushConst(sub), upvalues)
    case parser.ListLiteral:
        for _, x := range expr.List {
            err := comp.CompileExpr(x)
            if err != nil { return err }
        }
        comp.Frame.Write(interpreter.OP_MAKE_LIST, uint16(len(expr.List)))
    case parser.TableLiteral:
        for _, kv := range expr.Items {
            err := comp.CompileExpr(kv.Key)
            if err != nil { return err }
            err = comp.CompileExpr(kv.Value)
            if err != nil { return err }
        }
        comp.Frame.Write(interpreter.OP_MAKE_TABLE, uint16(len(expr.Items)))
    case parser.Symbol:
        sym := comp.VM.Symbol(expr.Value.Value)
        comp.Frame.Write(interpreter.OP_SYM, uint16(sym))
    case parser.Name:
        name := expr.Value.Lexeme

        if name == "__line__" {
            comp.Frame.Write(interpreter.OP_CONST, comp.PushConst(interpreter.Number(expr.Value.Line)))
            return nil
        }

        err := findName(comp, name)
        if err != nil {
            return fmt.Errorf("Line %d: %s", expr.Value.Line, err.Error())
        }
    case parser.Field:
        name := expr.Value.Value
        sym := comp.VM.Symbol(name)

        comp.Frame.Write(interpreter.OP_LOAD_FIELD, uint16(sym))
        return nil
    case parser.BasicLiteral:
        return comp.CompileLiteral(expr.Kind, expr.Value)
    case parser.StringInterpExpr:
        if len(expr.Parts) == 0 {
            comp.Frame.Write(interpreter.OP_CONST, comp.PushConst(interpreter.String("")))
            return nil
        }

        if len(expr.Parts) == 1 {
            comp.CompileExpr(expr.Parts[0])
            return nil
        }

        comp.Frame.Write(interpreter.OP_CONST, comp.PushConst(interpreter.String("")))
        for _, part := range expr.Parts {
            err := comp.CompileExpr(part)
            if err != nil { return err }
        }
        comp.Frame.Write(interpreter.OP_MAKE_LIST, uint16(len(expr.Parts)))
        comp.Frame.Write(interpreter.OP_SYM, uint16(comp.VM.Symbol("join:")))
        comp.Frame.Write(interpreter.OP_CALL, 1)
    }
    return nil
}

func (comp *CompilerState) CompileLiteral(tk token.Token, value string) error {
    switch tk.Kind {
    case token.NUMBER:
        num, err := strconv.ParseFloat(value, 64)
        if err != nil { return err }

        comp.Frame.Write(interpreter.OP_CONST, comp.PushConst(interpreter.Number(num)))
    case token.RAW_STRING, token.STRING_LITERAL:
        comp.Frame.Write(interpreter.OP_CONST, comp.PushConst(interpreter.String(value)))
    case token.REGEX:
        re, err := regexp.Compile(value)
        if err != nil { return err }

        comp.Frame.Write(interpreter.OP_CONST, comp.PushConst(interpreter.Regex{re}))
    }
    return nil
}

func validateParams(params []parser.Name) ([]string, error) {
    names := make([]string, len(params))

    for i, param := range params {
        name := param.Value.Value
        names[i] = name
        for j := 0; j < i; j++ {
            if names[j] == name {
                return names, fmt.Errorf("Repeated parameter name")
            }
        }
    }

    return names, nil
}

func validateSelector(selector parser.Expr) (symbol string, names []string, err error) {
    switch sel := selector.(type) {
    case parser.UnaryExpr:
        symbol = sel.Method.Value.Lexeme
        names = append(names,
            sel.X.(parser.Name).Value.Lexeme)
    case parser.BinaryExpr:
        symbol = sel.Op.Op.Lexeme
        names = append(names,
            sel.X.(parser.Name).Value.Lexeme,
            sel.Y.(parser.Name).Value.Lexeme)
    case parser.CallExpr:
        names = append(names, sel.Recv.(parser.Name).Value.Value)
        for i, param := range sel.Args {
            key := param.Key.Value.Value
            name := param.Value.(parser.Name).Value.Value
            names = append(names, name)
            for j := 0; j < i; j++ {
                if names[j] == name {
                    err = fmt.Errorf("Repeated parameter name")
                    return
                }
            }
            symbol += key
        }
    default:
        err = fmt.Errorf("Invalid method parameters")
    }

    return
}

func addUpvalue(comp *CompilerState, frame *StackFrame, index uint16, isLocal bool) int {
    count := len(comp.Frame.Upvalues)

    for i, upvalue := range frame.Upvalues {
        if upvalue.Index == index && upvalue.IsLocal == isLocal {
            return i
        }
    }
    
    frame.Upvalues = append(frame.Upvalues, Upvalue {
        Index: index,
        IsLocal: isLocal,
    })

    frame.Sub.UpvalueCount = uint16(count+1)

    return count
}

func findUpvalue(comp *CompilerState, frame *StackFrame, name string) int {
    if frame.Last == nil {
        return -1
    }

    loc, ok := frame.Last.Environment[name]
    if ok {
        loc.IsCaptured = true
        frame.Last.Environment[name] = loc
        return addUpvalue(comp, frame, loc.Slot, true)
    }

    upvalue := findUpvalue(comp, frame.Last, name)
    if upvalue != -1 {
        loc.IsCaptured = true
        frame.Last.Environment[name] = loc
        return addUpvalue(comp, frame, loc.Slot, false)
    }

    return -1
}

func findName(comp *CompilerState, name string) error {
    loc, ok := comp.Frame.Environment[name]
    if ok {
        comp.Frame.Write(interpreter.OP_LOAD_LOCAL, loc.Slot)
        return nil
    }

    upvalue := findUpvalue(comp, comp.Frame, name)
    if upvalue != -1 {
        comp.Frame.Write(interpreter.OP_LOAD_UPVALUE, uint16(upvalue))
        return nil
    }

    sym := comp.VM.Symbol(name)
    _, ok = comp.VM.TopModule.Table[sym]
    if !ok {
        return fmt.Errorf("No such name %s", name)
    }
    comp.Frame.Write(interpreter.OP_LOAD_MODULE, uint16(sym))
    return nil
}