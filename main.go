package main

import (
	"fmt"
	"os"

	"github.com/mattrco/difftrace/parser"
)

func main() {
	p := parser.NewParser(os.Stdin)
	for {
		line, err := p.Parse()
		if err != nil {
			if err != parser.ErrEOF {
				fmt.Println(err.Error())
			}
			break
		} else {
			fmt.Println("-------------------------\n")
			fmt.Printf("Signal: %v\n", line.Signal)
			fmt.Printf("FuncName: %v\n", line.FuncName)
			fmt.Printf("Args: %v\n", line.Args)
			for j, arg := range line.Args {
				fmt.Printf("\narg %v: %v\n", j, arg)
			}
			fmt.Printf("Result: %v\n", line.Result)
			fmt.Println(line.Unparse())
			fmt.Println("-------------------------\n")
		}
	}
}
