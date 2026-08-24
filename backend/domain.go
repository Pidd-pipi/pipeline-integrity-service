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
// cycleTransitionsActive 定义 active 状态允许的目标状态。
var cycleTransitionsActive = []string{}

// cycleTransitionsAccepted 定义 accepted 状态允许的目标状态。
var cycleTransitionsAccepted = []string{"active"}

// cycleTransitionTable 定义检验周期允许的状态迁移。
var cycleTransitionTable = map[string][]string{
	"planned":  {"active", "accepted"},
	"active":   cycleTransitionsActive,
	"accepted": cycleTransitionsAccepted,
}
