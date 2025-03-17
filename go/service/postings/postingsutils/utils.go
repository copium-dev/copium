package postingsutils

import (
	"fmt"
	"net/http"
	"strings"
	"strconv"
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
	location := params.Get("location")
	title := params.Get("title")
	startDate := params.Get("startDate")
	endDate := params.Get("endDate")

	// build filters for non free‑text filtering
	filters := make(map[string]string)
	if company != "" {
		filters["company_name"] = quoteIfNeeded(company)
	}
	if title != "" {
		filters["title"] = quoteIfNeeded(title)
	}
	if location != "" {
		filters["locations"] = quoteIfNeeded(location)
	}

	var filterStrs []string
	for key, value := range filters {
		filterStrs = append(filterStrs, fmt.Sprintf("%s:%s", key, value))
	}

	filtersString := strings.Join(filterStrs, " AND ")

	// NOTE: frontend sends unix time in ms but postings stores in seconds
	// so, we have to atoi, divide by 1000, then convert back to string
	// add date range filters if provided
	if startDate != "" && endDate != "" {
		startDateInt, err := strconv.Atoi(startDate)
		if err != nil {
			return "", "", err
		}
		endDateInt, err := strconv.Atoi(endDate)
		if err != nil {
			return "", "", err
		}
		startDate = fmt.Sprintf("%d", startDateInt / 1000)
		endDate = fmt.Sprintf("%d", endDateInt / 1000)
		filtersString = filtersString + fmt.Sprintf(" AND (date_updated >= %s AND date_updated <= %s)", startDate, endDate)
	} else if startDate != "" {
		startDateInt, err := strconv.Atoi(startDate)
		if err != nil {
			return "", "", err
		}
		startDate = fmt.Sprintf("%d", startDateInt / 1000)
		filtersString = filtersString + fmt.Sprintf(" AND date_updated >= %s", startDate)
	} else if endDate != "" {
		endDateInt, err := strconv.Atoi(endDate)
		if err != nil {
			return "", "", err
		}
		endDate = fmt.Sprintf("%d", endDateInt / 1000)
		filtersString = filtersString + fmt.Sprintf(" AND date_updated <= %s", endDate)
	}

	// prevent leading "AND"
	if len(filtersString) > 5 && filtersString[:5] == " AND " {
		filtersString = filtersString[5:]
	}

	return queryText, filtersString, nil
}

func quoteIfNeeded(val string) string {
	if strings.Contains(val, " ") && !(strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) {
		return fmt.Sprintf("\"%s\"", val)
	}
	return val
}
