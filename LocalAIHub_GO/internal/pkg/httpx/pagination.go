package httpx

import (
	"net/http"
	"strconv"
)

func ParsePage(r *http.Request, fallback int) int {
	value := r.URL.Query().Get("page")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func ParsePageSize(r *http.Request, fallback int) int {
	value := r.URL.Query().Get("page_size")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 || parsed > 200 {
		return fallback
	}
	return parsed
}

func Paginate(total, page, pageSize int) (int, int) {
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}
