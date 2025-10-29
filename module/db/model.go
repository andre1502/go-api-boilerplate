package db

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type TableTotalRows struct {
	TotalRows uint64 `gorm:"type:bigint UNSIGNED;column:total_rows" json:"-"`
}

type Sorting struct {
	OrderBy      *string `json:"order_by" query:"order_by"`
	MultiOrderBy *string
	Desc         *bool `json:"desc" query:"desc"`
}

type TableTime struct {
	TableCreatedAt
	TableUpdatedAt
}

type TableCreatedAt struct {
	CreatedAt *time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"created_at,omitempty,omitzero"`
}

type TableUpdatedAt struct {
	UpdatedAt *time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP" json:"updated_at,omitempty,omitzero"`
}

type TableDeleted struct {
	DeletedAt *time.Time            `gorm:"column:deleted_at;type:datetime" json:"-"`
	IsDeleted soft_delete.DeletedAt `gorm:"column:is_deleted;type:int;softDelete:flag,DeletedAtField:DeletedAt" json:"-"`
}

type TableUser struct {
	TableCreatedBy
	TableUpdatedBy
}

type TableCreatedBy struct {
	CreatedBy *string `gorm:"column:created_by;type:varchar(50);default:'System'" json:"created_by,omitempty,omitzero"`
}

type TableUpdatedBy struct {
	UpdatedBy *string `gorm:"column:updated_by;type:varchar(50);default:'System'" json:"updated_by,omitempty,omitzero"`
}
