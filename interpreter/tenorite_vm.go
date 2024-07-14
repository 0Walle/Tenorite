package interpreter

type TenoriteVM struct {
    SymbolStore   map[Symbol]string
    Modules       map[string]*Module
    TopModule     *Module
    Error         error

    StackTrace    bool
}

func MakeVM() TenoriteVM {
    var vm = TenoriteVM {
        SymbolStore: make(map[Symbol]string, 16),
        Modules: make(map[string]*Module, 1),
    }
    return vm
}

func (vm *TenoriteVM) Symbol(name string) Symbol {
    sym := makeSymbol(name)
    for {
        val, ok := vm.SymbolStore[sym]
        if val == name { return sym }
        if ok { sym += 1 }
        break
    }
    vm.SymbolStore[sym] = name
    return sym
}

func (vm *TenoriteVM) SymbolName(hash uint16) string {
    return vm.SymbolStore[Symbol(hash)]
}

func (vm *TenoriteVM) NewModule(name string) *Module {
    mod := &Module {
        Name: name,
        Variables: nil,
        Table: make(map[Symbol]int, 0),
    }
    vm.Modules[name] = mod
    return mod
}