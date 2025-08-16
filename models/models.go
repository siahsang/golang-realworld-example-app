package models

import "time"

type Profile struct {
	ID        int64   `json:"-"`
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

type Article struct {
	ID          int64     `json:"-"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AuthorID    int64     `json:"-"`
}

type Tag struct {
	ID   int64  `json:"-"`
	Name string `json:"name"`
}
