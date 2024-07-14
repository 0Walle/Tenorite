package main

import (
    "os"
    "io"
    "0Walle/Tenorite/compiler"
    
    "flag"
)

var file = flag.String("i", "", "input file")

func main() {
    flag.Parse()
    if *file == "" {
        fmt.Printf("No input file provided\n")
        return
    }

    f, err := os.Open(*file)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        return
    }

    source, _ := io.ReadAll(f)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        return
    }

    compiler.Compile(string(source)+"\n", *file)
}