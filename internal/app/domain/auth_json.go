package domain

type ctxKey string

const UserCtxKey ctxKey = "user"

type User struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type AuthorizationData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
