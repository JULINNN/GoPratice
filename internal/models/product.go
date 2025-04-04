package models

type Product struct {
	ID         int    `json:"id,omitempty" db:"id"`
	SkuCode    string `json:"sku_code,omitempty" db:"sku_code"`
	SkuName    string `json:"sku_name,omitempty" db:"sku_name"`
	SkuAmount  int    `json:"sku_amount,omitempty" db:"sku_amount"`
	Expiration string `json:"expiration,omitempty" db:"expiration"`
	CreateAt   string `json:"create_at,omitempty" db:"create_at"`
	UpdateAt   string `json:"update_at,omitempty" db:"update_at"`
}
