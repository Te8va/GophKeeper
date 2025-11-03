package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/Te8va/GophKeeper/internal/client/domain"
)

func createData() {
	sc := bufio.NewScanner(os.Stdin)
	CreateDataWithScanner(sc)
}

func CreateDataWithScanner(sc *bufio.Scanner) {
	fmt.Println("\n=== Создание записи ===")

	dataType := ChooseDataTypeWithScanner(sc)
	if dataType == "" {
		return
	}

	fmt.Print("Введите описание (метаданные): ")
	if !sc.Scan() {
		fmt.Println("Ошибка при считывании описания:", sc.Err())
		return
	}
	metadata := sc.Text()

	fmt.Print("Введите данные (текст/JSON): ")
	if !sc.Scan() {
		fmt.Println("Ошибка при считывании данных:", sc.Err())
		return
	}
	dataContent := sc.Text()

	dataReq := domain.CreateDataRequest{
		Type:     dataType,
		Metadata: metadata,
		Data:     dataContent,
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(dataReq); err != nil {
		fmt.Println("Ошибка при кодировании JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", domain.BaseURL+"/data", &buf)
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		fmt.Printf("Ошибка создания записи: %s - %s\n", response.Status, string(body))
		return
	}

	fmt.Println("✅ Запись успешно создана!")
}

func chooseDataType() domain.DataType {
	sc := bufio.NewScanner(os.Stdin)
	return ChooseDataTypeWithScanner(sc)
}

func ChooseDataTypeWithScanner(sc *bufio.Scanner) domain.DataType {
	fmt.Println("Выберите тип данных:")
	fmt.Println("1 - Логин/Пароль")
	fmt.Println("2 - Текстовые данные")
	fmt.Println("3 - Бинарные данные")
	fmt.Println("4 - Банковская карта")
	fmt.Print("Ваш выбор: ")

	if !sc.Scan() {
		fmt.Println("Ошибка при считывании выбора:", sc.Err())
		return ""
	}

	choice, err := strconv.Atoi(sc.Text())
	if err != nil {
		fmt.Println("Ошибка при преобразовании выбора:", err)
		return ""
	}

	switch choice {
	case 1:
		return domain.TypeLoginPassword
	case 2:
		return domain.TypeTextData
	case 3:
		return domain.TypeBinaryData
	case 4:
		return domain.TypeBankCard
	default:
		fmt.Println("Неверный выбор типа данных")
		return ""
	}
}

func getAllData() {
	fmt.Println("\n=== Все записи ===")

	req, err := http.NewRequest("GET", domain.BaseURL+"/data", nil)
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		fmt.Printf("Ошибка получения записей: %s - %s\n", response.Status, string(body))
		return
	}

	var records []domain.DataRecord
	if err := json.NewDecoder(response.Body).Decode(&records); err != nil {
		fmt.Println("Ошибка декодирования ответа:", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("Записей не найдено")
		return
	}

	fmt.Printf("✅ Найдено записей: %d\n", len(records))
	for i, record := range records {
		fmt.Printf("%d. ID: %s\n", i+1, record.ID)
		fmt.Printf("   Тип: %s\n", record.Type)
		fmt.Printf("   Описание: %s\n", record.Metadata)
		fmt.Printf("   Данные: %s\n", string(record.Data))
		fmt.Println("   ---")
	}
}

func getDataByID() {
	sc := bufio.NewScanner(os.Stdin)
	GetDataByIDWithScanner(sc)
}

func GetDataByIDWithScanner(sc *bufio.Scanner) {
	fmt.Println("\n=== Получить запись по ID ===")

	fmt.Print("Введите ID записи: ")
	if !sc.Scan() {
		fmt.Println("Ошибка при считывании ID:", sc.Err())
		return
	}
	id := sc.Text()

	req, err := http.NewRequest("GET", domain.BaseURL+"/data/"+id, nil)
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		fmt.Printf("Ошибка получения записи: %s - %s\n", response.Status, string(body))
		return
	}

	var record domain.DataRecord
	if err := json.NewDecoder(response.Body).Decode(&record); err != nil {
		fmt.Println("Ошибка декодирования ответа:", err)
		return
	}

	fmt.Println("✅ Запись найдена:")
	fmt.Printf("ID: %s\n", record.ID)
	fmt.Printf("Тип: %s\n", record.Type)
	fmt.Printf("Описание: %s\n", record.Metadata)
	fmt.Printf("Данные: %s\n", string(record.Data))
}

func updateData() {
	sc := bufio.NewScanner(os.Stdin)
	UpdateDataWithScanner(sc)
}

func UpdateDataWithScanner(sc *bufio.Scanner) {
	fmt.Println("\n=== Обновление записи ===")

	fmt.Print("Введите ID записи для обновления: ")
	if !sc.Scan() {
		fmt.Println("Ошибка при считывании ID:", sc.Err())
		return
	}
	id := sc.Text()
	if id == "" {
		fmt.Println("Ошибка при считывании ID: пустой ввод")
		return
	}

	dataType := ChooseDataTypeWithScanner(sc)
	if dataType == "" {
		return
	}

	fmt.Print("Введите новое описание: ")
	if !sc.Scan() {
		fmt.Println("Ошибка при считывании описания:", sc.Err())
		return
	}
	metadata := sc.Text()

	fmt.Print("Введите новые данные: ")
	if !sc.Scan() {
		fmt.Println("Ошибка при считывании данных:", sc.Err())
		return
	}
	dataContent := sc.Text()

	updateReq := map[string]interface{}{
		"type":     dataType,
		"metadata": metadata,
		"data":     dataContent,
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(updateReq); err != nil {
		fmt.Println("Ошибка при кодировании JSON:", err)
		return
	}

	req, err := http.NewRequest("PUT", domain.BaseURL+"/data/"+id, &buf)
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		fmt.Printf("Ошибка обновления записи: %s - %s\n", response.Status, string(body))
		return
	}

	fmt.Println("✅ Запись успешно обновлена!")
}

func deleteData() {
	sc := bufio.NewScanner(os.Stdin)
	DeleteDataWithScanner(sc)
}

func DeleteDataWithScanner(sc *bufio.Scanner) {
	fmt.Println("\n=== Удаление записи ===")

	fmt.Print("Введите ID записи: ")
	if !sc.Scan() {
		fmt.Println("Ошибка при считывании ID:", sc.Err())
		return
	}
	id := sc.Text()

	req, err := http.NewRequest("DELETE", domain.BaseURL+"/data/"+id, nil)
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		fmt.Printf("Ошибка удаления записи: %s - %s\n", response.Status, string(body))
		return
	}

	fmt.Println("✅ Запись успешно удалена!")
}

func SetToken(t string) {
	token = t
}

func GetToken() string {
	return token
}
