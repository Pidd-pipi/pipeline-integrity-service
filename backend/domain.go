package main

type IntegrityCycle struct {
	ID      string `json:"id"`
	Segment string `json:"segment"`
	Method  string `json:"method"`
	Risk    string `json:"risk"`
	Status  string `json:"status"`
	DueDate string `json:"dueDate"`
}
type StatusChange struct {
	Status string `json:"status"`
}
