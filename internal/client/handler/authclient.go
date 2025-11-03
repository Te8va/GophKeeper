package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Te8va/GophKeeper/internal/client/domain"
)

func Register() {
	fmt.Println("\n=== Регистрация ===")

	login, password := readCredentials()
	if login == "" || password == "" {
		return
	}

	authReq := domain.AuthRequest{
		Login:    login,
		Password: password,
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(authReq); err != nil {
		fmt.Println("Ошибка при кодировании JSON:", err)
		return
	}

	response, err := http.Post(domain.BaseURL+"/user/register", "application/json", &buf)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	defer response.Body.Close()

	for _, cookie := range response.Cookies() {
		if cookie.Name == "auth_token" {
			token = cookie.Value
			break
		}
	}

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		fmt.Printf("Ошибка регистрации: %s - %s\n", response.Status, string(body))
		return
	}

	fmt.Println("✅ Регистрация успешна!")
}

func Login() {
	fmt.Println("\n=== Логин ===")

	login, password := readCredentials()
	if login == "" || password == "" {
		return
	}

	authReq := domain.AuthRequest{
		Login:    login,
		Password: password,
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(authReq); err != nil {
		fmt.Println("Ошибка при кодировании JSON:", err)
		return
	}

	response, err := http.Post(domain.BaseURL+"/user/login", "application/json", &buf)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	defer response.Body.Close()

	for _, cookie := range response.Cookies() {
		if cookie.Name == "auth_token" {
			token = cookie.Value
			break
		}
	}

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		fmt.Printf("Ошибка логина: %s - %s\n", response.Status, string(body))
		return
	}

	fmt.Println("✅ Успешный вход в систему!")
}

func readCredentials() (string, string) {
	sc := bufio.NewScanner(os.Stdin)

	fmt.Print("Введите логин: ")
	if !sc.Scan() {
		fmt.Println("Ошибка при считывании логина:", sc.Err())
		return "", ""
	}
	login := sc.Text()

	fmt.Print("Введите пароль: ")
	if !sc.Scan() {
		fmt.Println("Ошибка при считывании пароля:", sc.Err())
		return "", ""
	}
	password := sc.Text()

	return login, password
}
