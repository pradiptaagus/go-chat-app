package domain

import "time"

type Client struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	CreatedDate time.Time `json:"createdDate"`
}
