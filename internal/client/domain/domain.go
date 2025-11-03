package domain

import (
	"encoding/json"
)

var (
	BaseURL    = "http://localhost:8080"
	ServerPort = "8080"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type DataType string

const (
	TypeLoginPassword DataType = "login_password"
	TypeTextData      DataType = "text_data"
	TypeBinaryData    DataType = "binary_data"
	TypeBankCard      DataType = "bank_card"
)

type CreateDataRequest struct {
	Type     DataType    `json:"type"`
	Metadata string      `json:"metadata"`
	Data     interface{} `json:"data"`
}

type DataRecord struct {
	ID       string          `json:"id"`
	UserID   string          `json:"user_id"`
	Type     DataType        `json:"type"`
	Data     json.RawMessage `json:"data"`
	Metadata string          `json:"metadata"`
}
