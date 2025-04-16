package requests

type ApiRequest struct {
	Action      string            `json:"method"`  // HTTP method (GET, POST, etc.)
	RequestParams map[string]string `json:"query_params,omitempty"`      // Optional query parameters
}

type CreateTaskRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateTaskRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	IsDone  bool   `json:"is_done,omitempty"`
}