package timeutil

import "fmt"

func GetFormattedDuration(totalMinutes int) string {
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
