package domain

type DataType string

const (
	TypeLoginPassword DataType = "login_password"
	TypeTextData      DataType = "text"
	TypeBinaryData    DataType = "binary"
	TypeBankCard      DataType = "bank_card"
)

type DataRecord struct {
	ID       string   `json:"id"`
	UserID   string   `json:"user_id"`
	Type     DataType `json:"type"`
	Metadata string   `json:"metadata"`
	Data     []byte   `json:"data"`
}

type CreateDataRequest struct {
	Type     DataType `json:"type"`
	Metadata string   `json:"metadata"`
	Data     any      `json:"data"`
}

type UpdateDataRequest struct {
	ID       string   `json:"id"`
	Type     DataType `json:"type"`
	Metadata string   `json:"metadata"`
	Data     any      `json:"data"`
}
