package main

import "fmt"

func validateCycleStatus(s string) error {
	switch s {
	case "planned", "active", "accepted":
		return nil
	default:
		return fmt.Errorf("status must be planned, active, or accepted")
	}
}
// validateCycleTransition 校验周期状态迁移是否允许。
// 先确认目标状态本身合法，再依据 cycleTransitionTable 校验 current→next 是否在允许范围内，
// 自跳转（current==next）一律视为非法，避免验收后反复横跳。
func validateCycleTransition(current, next string) error {
	if err := validateCycleStatus(next); err != nil {
		return err
	}
	if current == next || !cycleAllowsTransition(current, next) {
		return fmt.Errorf("status transition %s to %s is not allowed", current, next)
	}
	return nil
}
