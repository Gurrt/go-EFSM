package main

import (
	"EFSM"
	"fmt"
	"net/http"
)

func main() {
	eims, err := EFSM.FromJSONFile("definition.json")
	if err != nil {
		fmt.Print(err)
		return
	}
	go EFSM.StartAPI(eims)
	fmt.Println("Starting HTML Interface @ port 80")
	panic(http.ListenAndServe(":80", http.FileServer(http.Dir("html"))))
	/*
		for i := range eims {
			eims[i].Print(i)
		}
		for true {
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Enter index: ")
			indexStr, _ := reader.ReadString('\n')
			index64, _ := strconv.ParseInt(indexStr[0:len(indexStr)-2], 10, 0)
			index := int(index64)

			fmt.Print("Enter Function: ")
			text2, _ := reader.ReadString('\n')
			function := text2[0 : len(text2)-2]
			if function == "print" {
				eims[index].Print(index)
				continue
			}

			fmt.Print("Enter ID: ")
			text, _ := reader.ReadString('\n')
			id := text[0 : len(text)-2]

			ids := strings.Split(id, ",")
			for i := range ids {
				err := eims[index].Exec(strings.TrimSpace(ids[i]), function)

				if err != nil {
					fmt.Println(err)
				}
			}
		}
	*/
}
