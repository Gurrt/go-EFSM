package main

import (
	"EFSM"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	eim, err := EFSM.FromJSONFile("definition.json")
	if err != nil {
		fmt.Print(err)
		return
	}
	eim.Print()
	for true {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter Function: ")
		text2, _ := reader.ReadString('\n')
		function := text2[0 : len(text2)-2]
		if function == "print" {
			eim.Print()
			continue
		}

		fmt.Print("Enter ID: ")
		text, _ := reader.ReadString('\n')
		id := text[0 : len(text)-2]

		ids := strings.Split(id, ",")
		for i := range ids {
			err := eim.Exec(strings.TrimSpace(ids[i]), function)

			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
