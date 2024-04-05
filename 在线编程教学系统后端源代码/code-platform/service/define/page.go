package define

type PageInfo struct {
	Total int `json:"total"` // 总记录数
}

type PageResponse struct {
	Records  interface{} `json:"records"`
	PageInfo *PageInfo   `json:"page_info"`
}
