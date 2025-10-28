package model

import (
	"fmt"
	"go-api-boilerplate/module/db"
)

type User struct {
	ID int `gorm:"column:id;type:int;primaryKey;autoIncrement" json:"id"`
	db.TableTime
	db.TableUser
}

func (User) TableName() string {
	return "user"
}

func (m *User) Alias(name string) string {
	return fmt.Sprintf("%s AS %s", m.TableName(), name)
}
