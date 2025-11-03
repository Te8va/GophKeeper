package main

import (
	"fmt"

	"github.com/Te8va/GophKeeper/internal/client/handler"
)

func main() {
	fmt.Println("=== GophKeeper Client ===")

	for {
		action, err := handler.ReadUserChoice()
		if err != nil {
			fmt.Println("Ошибка:", err)
			continue
		}
		handler.ChoiceAction(action)
	}
}
