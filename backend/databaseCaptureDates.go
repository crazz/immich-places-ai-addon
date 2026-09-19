package main

// captureDaySQL reads the recorded calendar date without converting its offset.
// column is an internal SQL identifier, never request input.
func captureDaySQL(column string) string {
	day := "substr(" + column + ", 1, 10)"
	return "CASE WHEN date(" + day + ", '+0 days') = " + day + " THEN " + day + " END"
}

// captureRangeSQL preserves indexable timestamp bounds and inclusive end days.
// ISO source timestamps begin with their calendar date, independently of offset.
func captureRangeSQL(column, startDate, endDate string) (string, []interface{}) {
	if startDate == "" && endDate == "" {
		return "", nil
	}
	clause := " AND " + captureDaySQL(column) + " IS NOT NULL"
	var args []interface{}
	if startDate != "" {
		clause += " AND " + column + " >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		clause += " AND " + column + " < ?"
		args = append(args, endDate+"T99")
	}
	return clause, args
}
