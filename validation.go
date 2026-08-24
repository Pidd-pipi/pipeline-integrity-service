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
