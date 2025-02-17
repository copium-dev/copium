package userutils

import (
	"fmt"
	"net/http"
	"strings"
)

func CalculateTotalPages(totalHits int, hitsPerPage int) int {
	if hitsPerPage <= 0 {
		return 0
	}
	return (totalHits + hitsPerPage - 1) / hitsPerPage
}

func ParseQuery(r *http.Request) (string, string) {
	params := r.URL.Query()
	queryText := params.Get("q")
	company := params.Get("company")
	statusParam := params.Get("status")
	role := params.Get("role")
	location := params.Get("location")
	startDate := params.Get("startDate")
	endDate := params.Get("endDate")

	// build filters for non free‑text filtering
	filters := make(map[string]string)
	if company != "" {
		filters["company"] = company
	}
	if statusParam != "" {
		filters["status"] = statusParam
	}
	if role != "" {
		filters["role"] = role
	}
	if location != "" {
		filters["location"] = location
	}

	var filterStrs []string
	for key, value := range filters {
		filterStrs = append(filterStrs, fmt.Sprintf("%s:%s", key, value))
	}

	filtersString := strings.Join(filterStrs, " AND ")

	// add date range filters if provided
	if startDate != "" && endDate != "" {
		filtersString = filtersString + fmt.Sprintf(" AND (appliedDate >= %s AND appliedDate <= %s)", startDate, endDate)
	} else if startDate != "" {
		filtersString = filtersString + fmt.Sprintf(" AND appliedDate >= %s", startDate)
	} else if endDate != "" {
		filtersString = filtersString + fmt.Sprintf(" AND appliedDate <= %s", endDate)
	}

	// prevent leading "AND"
	if len(filtersString) > 5 && filtersString[:5] == " AND " {
		filtersString = filtersString[5:]
	}

	return queryText, filtersString
}