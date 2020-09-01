package model

type Message struct {
	Text       string `json:"text"`
	Action     string `json:"action"`
	RoomNumber string `json:"roomnumber"`
}
