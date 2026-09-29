package types

// LogPageReq is the query parameter struct for log list endpoints.
// Manually maintained (not goctl-generated) to avoid modifying the generated types.go.
type LogPageReq struct {
	Page         int    `json:"page,optional" form:"page,optional"`
	PageSize     int    `json:"pageSize,optional" form:"pageSize,optional"`
	Username     string `json:"username,optional" form:"username,optional"`
	Status       string `json:"status,optional" form:"status,optional"`
	Ip           string `json:"ip,optional" form:"ip,optional"`
	Module       string `json:"module,optional" form:"module,optional"`
	Operation    string `json:"operation,optional" form:"operation,optional"`
	OperatorName string `json:"operatorName,optional" form:"operatorName,optional"`
	Path         string `json:"path,optional" form:"path,optional"`
}
