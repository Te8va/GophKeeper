package handler

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

var (
	token  string
	osExit = os.Exit
)

func ReadUserChoice() (int, error) {
	fmt.Println("\n=== GophKeeper Menu ===")
	if token == "" {
		fmt.Println("1 - Регистрация")
		fmt.Println("2 - Логин")
	} else {
		fmt.Println("3 - Создать запись")
		fmt.Println("4 - Показать все записи")
		fmt.Println("5 - Получить запись по ID")
		fmt.Println("6 - Обновить запись")
		fmt.Println("7 - Удалить запись")
		fmt.Println("8 - Выйти из аккаунта")
	}
	fmt.Println("0 - Выход из программы")
	fmt.Print("Выберите команду: ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return -1, fmt.Errorf("ошибка при считывании ввода: %v", scanner.Err())
	}

	action, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return -1, fmt.Errorf("ошибка при преобразовании команды в число: %v", err)
	}

	return action, nil
}

func ChoiceAction(action int) {
	switch action {
	case 0:
		fmt.Println("Выход из программы.")
		osExit(0)
	case 1:
		Register()
	case 2:
		Login()
	case 3:
		if checkAuth() {
			createData()
		}
	case 4:
		if checkAuth() {
			getAllData()
		}
	case 5:
		if checkAuth() {
			getDataByID()
		}
	case 6:
		if checkAuth() {
			updateData()
		}
	case 7:
		if checkAuth() {
			deleteData()
		}
	case 8:
		logout()
	default:
		fmt.Println("Неверная команда:", action)
	}
}

func checkAuth() bool {
	if token == "" {
		fmt.Println("Ошибка: необходимо сначала войти в систему")
		return false
	}
	return true
}

func logout() {
	token = ""
	fmt.Println("✅ Вы вышли из аккаунта")
}
