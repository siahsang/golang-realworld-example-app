package models

type Profile struct {
	ID        int64   `json:"-"`
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}
