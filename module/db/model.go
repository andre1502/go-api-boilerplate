package db

import "time"

type TableTotalRows struct {
	TotalRows uint64 `gorm:"type:bigint UNSIGNED;column:total_rows" json:"-"`
}

type Sorting struct {
	OrderBy      *string `json:"order_by" query:"order_by"`
	MultiOrderBy *string
	Desc         *bool `json:"desc" query:"desc"`
}

type TableTime struct {
	CreatedAt *time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:新增時間" json:"created_at,omitempty"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:修改時間" json:"updated_at,omitempty"`
}

type TableCreatedAt struct {
	CreatedAt *time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:新增時間" json:"created_at,omitempty"`
}

type TableUpdatedAt struct {
	UpdatedAt *time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:修改時間" json:"updated_at,omitempty"`
}

type TableUser struct {
	CreatedBy *string `gorm:"column:created_by;type:varchar(50);comment:新增後台帳號" json:"created_by,omitempty"`
	UpdatedBy *string `gorm:"column:updated_by;type:varchar(50);comment:修改後台帳號" json:"updated_by,omitempty"`
}

type TableCreatedBy struct {
	CreatedBy *string `gorm:"column:created_by;type:varchar(50);comment:新增後台帳號" json:"created_by,omitempty"`
}

type TableUpdatedBy struct {
	UpdatedBy *string `gorm:"column:updated_by;type:varchar(50);comment:修改後台帳號" json:"updated_by,omitempty"`
}
