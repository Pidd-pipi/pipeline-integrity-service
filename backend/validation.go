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
func validateCycleTransition(current, next string) error {
	return validateCycleStatus(next)
}
