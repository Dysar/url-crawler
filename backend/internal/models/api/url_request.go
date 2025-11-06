package api

type CreateURLRequest struct {
    URL string `json:"url" binding:"required,url"`
}


