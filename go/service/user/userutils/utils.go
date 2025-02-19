package userutils

import (
	"fmt"
	"net/http"
	"strings"
	"slices"
)

func CalculateTotalPages(totalHits int, hitsPerPage int) int {
	if hitsPerPage <= 0 {
		return 0
	}
	return (totalHits + hitsPerPage - 1) / hitsPerPage
}

func ParseQuery(r *http.Request) (string, string, error) {
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
		filters["company"] = quoteIfNeeded(company)
	}
	if statusParam != "" {
		// validate status first 
		err := checkStatusParam(statusParam)
		if err != nil {
			// return 400 error
			return "", "", err
		}
		filters["status"] = statusParam	
	}
	if role != "" {
		filters["role"] = quoteIfNeeded(role)
	}
	if location != "" {
		filters["location"] = quoteIfNeeded(location)
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

	return queryText, filtersString, nil
}

// if filter contains a space, it needs to be quoted
// we need this to avoid below error
// Error: API error [400] filters:
// Unexpected token string(intern) expected end of filter at col 9
func quoteIfNeeded(val string) string {
    if strings.Contains(val, " ") && !(strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) {
        return fmt.Sprintf("\"%s\"", val)
    }
    return val
}

// frontend only displays a dropdown for status filtering
// but some clever users can pass in a different value
func checkStatusParam(val string) error {
	validStatuses := []string{"Applied", "Screen", "Interviewing", "Offer", "Rejected", "Ghosted"}
	if !slices.Contains(validStatuses, val) {
		// return 400 error
		return fmt.Errorf("Invalid status: %s", val)
	}

	return nil
}