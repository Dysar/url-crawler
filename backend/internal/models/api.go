package models

type CreateURLRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type URLResponse struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}
