package web

import "time"

type CreateClientResponse struct {
	UserID      int64     `json:"userId"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	CreatedDate time.Time `json:"createdDate"`
}
