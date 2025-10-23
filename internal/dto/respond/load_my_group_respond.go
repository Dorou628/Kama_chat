package respond

type LoadMyGroupRespond struct {
	GroupId   string `json:"group_id"`
	GroupName string `json:"name"`
	Avatar    string `json:"avatar"`
}
