package main

import (
	"EFSM"
	"bufio"
	"fmt"
	"os"
)

func main() {
	efsm, err := EFSM.FromJSONFile("definition.json")
	if err != nil {
		fmt.Print(err)
	} else {
		for true {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter function: ")
			text, _ := reader.ReadString('\n')
			text = text[0 : len(text)-2]
			if text == "print" {
				fmt.Println("")
				efsm.Print()
				fmt.Println("")
			} else {
				efsm.ExecuteFunction(text)
				fmt.Println("New state: ", efsm.GetCurrentStateName())
			}
		}
	}
}
