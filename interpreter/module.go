package interpreter

type Module struct {
    Name       string
    Variables  []Receiver
    Table      map[Symbol]int
}

func (mod *Module) Add(name Symbol, value Receiver) {
    loc := len(mod.Variables)
    mod.Variables = append(mod.Variables, value)
    mod.Table[name] = loc
}

func (mod *Module) Reserve(name Symbol) {
    loc := len(mod.Variables)
    mod.Variables = append(mod.Variables, nil)
    mod.Table[name] = loc
}

func (mod *Module) Get(name Symbol) (Receiver, bool) {
    loc, ok := mod.Table[name]
    if !ok {
        return nil, false
    }
    return mod.Variables[loc], true
}