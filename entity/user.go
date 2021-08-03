package entity

//Person object for REST(CRUD)
type User struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Balance       float64 `json:"balance"`
	Created_time  string  `json:"created_time"`
	Modified_time string  `json:"modified_time"`
}

type ChangeBalance struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	TargetId int64   `json:"targetId"`
}
