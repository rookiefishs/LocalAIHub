package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func DecodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}
