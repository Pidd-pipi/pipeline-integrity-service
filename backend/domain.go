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
// cycleTransitionsActive 定义 active 状态允许的目标状态：周期进入进行中后只能被验收。
var cycleTransitionsActive = []string{"accepted"}

// cycleTransitionsAccepted 定义 accepted 状态允许的目标状态：验收即终态，不可再流转。
var cycleTransitionsAccepted = []string{}

// cycleTransitionTable 定义检验周期允许的状态迁移。
// 生命周期为 planned → active → accepted，accepted 为终态；任何自跳转都不允许。
var cycleTransitionTable = map[string][]string{
	"planned":  {"active", "accepted"},
	"active":   cycleTransitionsActive,
	"accepted": cycleTransitionsAccepted,
}

// cycleAllowsTransition 报告 current→next 是否为合法的周期状态迁移。
func cycleAllowsTransition(current, next string) bool {
	for _, allowed := range cycleTransitionTable[current] {
		if allowed == next {
			return true
		}
	}
	return false
}
