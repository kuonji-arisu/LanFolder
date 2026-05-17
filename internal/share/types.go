package share

import (
	"time"

	"lanfolder/i18n"
)

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

func PermissionOptions(language string) []PermissionOption {
	return []PermissionOption{
		{Value: PermissionReadOnly, Label: i18n.T(language, "permission.readonly.label", nil), Description: i18n.T(language, "permission.readonly.description", nil)},
		{Value: PermissionUpload, Label: i18n.T(language, "permission.upload.label", nil), Description: i18n.T(language, "permission.upload.description", nil)},
		{Value: PermissionManage, Label: i18n.T(language, "permission.manage.label", nil), Description: i18n.T(language, "permission.manage.description", nil)},
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
