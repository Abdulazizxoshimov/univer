package models


type (
	CategoryReq struct {
		Name string `json:"category_name"`
	}

	Category struct {
		ID   string `json:"category_id"`
		Name string `json:"category_name"`
	}

	ListCategory struct {
		Categories []Category `json:"categories"`
		Total      uint64     `json:"total_count"`
	}
)