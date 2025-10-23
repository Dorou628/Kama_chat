package respond

type LoadMyJoinedGroupRespond struct {
	GroupId   string `json:"group_id"`
	GroupName string `json:"name"`
	Avatar    string `json:"avatar"`
}
