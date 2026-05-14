package share

import "time"

type Permission string

const (
	PermissionReadOnly Permission = "readonly"
	PermissionUpload   Permission = "upload"
	PermissionManage   Permission = "manage"
)

func (p Permission) CanUpload() bool {
	return p == PermissionUpload || p == PermissionManage
}

func (p Permission) CanDelete() bool {
	return p == PermissionManage
}

func (p Permission) Valid() bool {
	return p == PermissionReadOnly || p == PermissionUpload || p == PermissionManage
}

type PermissionOption struct {
	Value       Permission `json:"value"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
}

func PermissionOptions() []PermissionOption {
	return []PermissionOption{
		{Value: PermissionReadOnly, Label: "只读", Description: "浏览和下载"},
		{Value: PermissionUpload, Label: "可上传", Description: "浏览、下载、上传"},
		{Value: PermissionManage, Label: "可删改", Description: "浏览、下载、上传、删除"},
	}
}

type Entry struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"isDir"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"modTime"`
	Extension string    `json:"extension,omitempty"`
}

type ListResult struct {
	Path       string  `json:"path"`
	ParentPath string  `json:"parentPath"`
	Entries    []Entry `json:"entries"`
}

type Message struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	ClientID  string    `json:"clientId"`
	Text      string    `json:"text"`
}

type Status struct {
	Root       string     `json:"root"`
	Permission Permission `json:"permission"`
}
