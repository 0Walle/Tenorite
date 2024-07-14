package interpreter

// import (
//     "fmt"
// )

// func PrintLine(line int, index int, op uint16, arg uint16, fstr string, other ...interface{}) {
//     fmt.Printf(" %-4d %6d %-14s", line, index, OPCODE_NAMES[op])
//     if arg < 256 {
//         fmt.Printf("%4d ", arg)
//     } else if arg != 0xffff {
//         fmt.Printf("%04x ", arg)
//     }
//     fmt.Printf(fstr, other...)
//     fmt.Printf("\n")
// }

// func PrintSub(ctx *VMState, sub *CodeObj) {
//     fmt.Printf("<code object %s> [ ", sub.Name)
//     for _, name := range sub.CoVarnames {
//         fmt.Printf("%#v ", name)
//     }
//     fmt.Printf("] (1+%d+%d)\n", sub.Arity, sub.LocalSize)
//     // fmt.Printf("%#v\n", sub.Lines)
//     ip := 0
//     line := 1
//     for {

//         if len(sub.Lines) > 0 && ip <= sub.Lines[line] {
//             for line+1 < len(sub.Lines) && ip >= sub.Lines[line] {
//                 line += 1
//             }
//         }

//         op := sub.Code[ip]
//         switch op {
//         case OP_NOP,
//              OP_POP,
//              OP_RETURN,
//              OP_RECURSIVE:
//             PrintLine(line, ip, op, 0xffff, "")
//             ip+=1
        
//         case OP_CONST:
//             at := sub.Code[ip+1]
//             PrintLine(line, ip, op, at, "(%v)", toString(ctx, sub.Consts[at]))
//             ip+=2
            
//         case OP_SYM,
//              OP_MESSAGE:
//             sym := sub.Code[ip+1]
//             PrintLine(line, ip, op, sym, "(%s)", ctx.SymbolName(sym))
//             ip+=2

//         case OP_CLOSURE:
//             at := sub.Code[ip+1]
//             PrintLine(line, ip, op, at, "(%v)", toString(ctx, sub.Consts[at]))
//             count := int(sub.Consts[at].(*CodeObj).UpvalueCount)
//             for i := 0; i < count; i++ {
//                 ip+=2
//                 isLocal := sub.Code[ip]
//                 index := sub.Code[ip+1]
//                 if isLocal != 0 {
//                     PrintLine(line, ip, op, index, "(^%v)", sub.CoVarnames[index])
//                 } else {
//                     PrintLine(line, ip, op, index, "(^%v)", index)
//                 }
                
//             }
//             ip+=2
//         case OP_STORE_MODULE, OP_LOAD_MODULE:
//             sym := sub.Code[ip+1]
//             PrintLine(line, ip, op, sym, "(%s)", ctx.SymbolName(sym))
//             ip+=2
        
//         case OP_STORE_LOCAL, OP_LOAD_LOCAL:
//             at := sub.Code[ip+1]
//             PrintLine(line, ip, op, at, "(%s)", sub.CoVarnames[at])
//             ip+=2

//         case OP_STORE_UPVALUE, OP_LOAD_UPVALUE:
//             at := sub.Code[ip+1]
//             PrintLine(line, ip, op, at, "")
//             ip+=2

//         case OP_OPERATOR:
//             sym := sub.Code[ip+1]
//             PrintLine(line, ip, op, sym, "(%s)", ctx.SymbolName(sym))
//             ip+=2

//         case OP_CALL_0,
//              OP_CALL_1,
//              OP_CALL_2,
//              OP_CALL_3,
//              OP_CALL_4:
//             sym := sub.Code[ip+1]
//             PrintLine(line, ip, op, sym, "(%s)", ctx.SymbolName(sym))
//             ip+=2

//         case OP_CALL_X:
//             nargs := sub.Code[ip+1]
//             PrintLine(line, ip, op, nargs, "")
//             ip+=2
//         case OP_JUMP_FALSE:
//             forward := sub.Code[ip+1]
//             PrintLine(line, ip, op, forward, "jump to %d", ip+int(forward))
//             ip+=2
//         case OP_LOOP:
//             backward := sub.Code[ip+1]
//             PrintLine(line, ip, op, backward, "jump to %d", ip-int(backward))
//             ip+=2
//         case OP_END:
//             return
//             // PrintLine(line, ip, "END", 0xffff, "")

//         case OP_CLOSE_UPVALUE:
//             at := sub.Code[ip+1]
//             PrintLine(line, ip, op, at, "(%s)", sub.CoVarnames[at])
//             ip+=2
//         case OP_MAKE_LIST,
//              OP_MAKE_TABLE:
//             n := sub.Code[ip+1]
//             PrintLine(line, ip, op, n, "")
//             ip+=2
//         case OP_MAKE_METHOD:
//             sym := sub.Code[ip+1]
//             PrintLine(line, ip, op, sym, "(%s)", ctx.SymbolName(sym))
//             ip+=2
//         default:
//             panic("Invalid OPCODE")
//             return
//         }
//     }
// }