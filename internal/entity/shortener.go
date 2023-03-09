// Structs for shortener app.

package entity

// URLs struct for history response.
type URLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// URLBatch struct for batch request.
type URLBatch struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
}

// ReqJSON struct for single application/json request.
type ReqJSON struct {
	URL string `json:"url"`
}

// RespJSON struct for single application/json response.
type RespJSON struct {
	Result string `json:"result"`
}
