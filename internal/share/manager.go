package share

import (
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const trashDirName = "trash"

type Manager struct {
	mu         sync.RWMutex
	root       string
	rootReal   string
	permission Permission
	showHidden bool
	messages   *MessageStore
}

type ResolvedPath struct {
	CleanRel string
	UserPath string
	RealPath string
	Info     os.FileInfo
}

func NewManager() *Manager {
	return &Manager{permission: PermissionReadOnly, messages: NewMessageStore()}
}

func (m *Manager) Configure(root string, permission Permission, showHidden bool) error {
	if !permission.Valid() {
		permission = PermissionReadOnly
	}
	if root == "" {
		return ErrInvalidRoot
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return ErrInvalidRoot
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.root = abs
	m.rootReal = real
	m.permission = permission
	m.showHidden = showHidden
	return nil
}

func (m *Manager) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Status{Root: m.root, Permission: m.permission}
}

func (m *Manager) resolveExisting(rel string) (ResolvedPath, error) {
	cleaned, err := cleanRel(rel)
	if err != nil {
		return ResolvedPath{}, err
	}
	m.mu.RLock()
	root := m.root
	rootReal := m.rootReal
	showHidden := m.showHidden
	m.mu.RUnlock()
	if root == "" || rootReal == "" {
		return ResolvedPath{}, ErrInvalidRoot
	}
	if isManagedPath(cleaned) {
		return ResolvedPath{}, ErrInvalidPath
	}
	if !showHidden && containsHiddenSegment(cleaned) {
		return ResolvedPath{}, ErrNotFound
	}
	target := filepath.Join(root, filepath.FromSlash(cleaned))
	real, err := filepath.EvalSymlinks(target)
	if os.IsNotExist(err) {
		return ResolvedPath{}, ErrNotFound
	}
	if err != nil {
		return ResolvedPath{}, err
	}
	if !withinRoot(rootReal, real) {
		return ResolvedPath{}, ErrPathEscape
	}
	info, err := os.Stat(real)
	if err != nil {
		return ResolvedPath{}, err
	}
	return ResolvedPath{
		CleanRel: cleaned,
		UserPath: target,
		RealPath: real,
		Info:     info,
	}, nil
}

func (m *Manager) resolveDirectory(rel string) (ResolvedPath, error) {
	resolved, err := m.resolveExisting(rel)
	if err != nil {
		return ResolvedPath{}, err
	}
	if !resolved.Info.IsDir() {
		return ResolvedPath{}, ErrInvalidPath
	}
	return resolved, nil
}

func (m *Manager) List(rel string) (ListResult, error) {
	resolved, err := m.resolveDirectory(rel)
	if err != nil {
		return ListResult{}, err
	}
	items, err := os.ReadDir(resolved.RealPath)
	if err != nil {
		return ListResult{}, err
	}

	m.mu.RLock()
	showHidden := m.showHidden
	m.mu.RUnlock()

	entries := make([]Entry, 0, len(items))
	for _, item := range items {
		name := item.Name()
		entryPath := path.Join(resolved.CleanRel, name)
		if isManagedPath(slashRel(entryPath)) {
			continue
		}
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(resolved.RealPath, name)
		real, err := filepath.EvalSymlinks(full)
		if err != nil || !m.isWithinCurrentRoot(real) {
			continue
		}
		info, err := os.Stat(real)
		if err != nil {
			continue
		}
		entries = append(entries, Entry{
			Name:      name,
			Path:      slashRel(entryPath),
			IsDir:     info.IsDir(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			Extension: strings.TrimPrefix(filepath.Ext(name), "."),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	parent := ""
	if resolved.CleanRel != "" {
		parent = path.Dir(resolved.CleanRel)
		if parent == "." {
			parent = ""
		}
	}
	return ListResult{Path: resolved.CleanRel, ParentPath: parent, Entries: entries}, nil
}

func (m *Manager) OpenForDownload(rel string) (*os.File, Entry, error) {
	resolved, err := m.resolveExisting(rel)
	if err != nil {
		return nil, Entry{}, err
	}
	if resolved.Info.IsDir() {
		return nil, Entry{}, ErrInvalidPath
	}
	file, err := os.Open(resolved.RealPath)
	if err != nil {
		return nil, Entry{}, err
	}
	return file, Entry{Name: resolved.Info.Name(), Path: resolved.CleanRel, Size: resolved.Info.Size(), ModTime: resolved.Info.ModTime()}, nil
}

func (m *Manager) SaveUpload(rel string, header *multipart.FileHeader) (Entry, error) {
	m.mu.RLock()
	canUpload := m.permission.CanUpload()
	m.mu.RUnlock()
	if !canUpload {
		return Entry{}, ErrPermissionDenied
	}
	resolved, err := m.resolveDirectory(rel)
	if err != nil {
		return Entry{}, err
	}
	name, err := cleanFilename(header.Filename)
	if err != nil {
		return Entry{}, err
	}
	src, err := header.Open()
	if err != nil {
		return Entry{}, err
	}
	defer src.Close()
	out, createdName, err := createUniqueFile(resolved.RealPath, name, 0644)
	if err != nil {
		return Entry{}, err
	}
	dst := filepath.Join(resolved.RealPath, createdName)
	_, copyErr := io.Copy(out, src)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return Entry{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(dst)
		return Entry{}, closeErr
	}
	info, err := os.Stat(dst)
	if err != nil {
		return Entry{}, err
	}
	relPath := path.Join(resolved.CleanRel, createdName)
	return Entry{Name: info.Name(), Path: slashRel(relPath), Size: info.Size(), ModTime: info.ModTime()}, nil
}

func (m *Manager) Delete(rel string) error {
	m.mu.RLock()
	canDelete := m.permission.CanDelete()
	root := m.root
	m.mu.RUnlock()
	if !canDelete {
		return ErrPermissionDenied
	}
	cleaned, err := cleanRel(rel)
	if err != nil {
		return err
	}
	if cleaned == "" {
		return ErrCannotDeleteRoot
	}
	if isManagedPath(cleaned) {
		return ErrInvalidPath
	}
	resolved, err := m.resolveExisting(cleaned)
	if err != nil {
		return err
	}
	trashDir := filepath.Join(root, managedDirName, trashDirName)
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return err
	}
	name := trashName(resolved.CleanRel)
	dst := uniquePath(trashDir, name)
	return os.Rename(resolved.UserPath, dst)
}

func (m *Manager) Mkdir(rel, name string) (Entry, error) {
	m.mu.RLock()
	canUpload := m.permission.CanUpload()
	m.mu.RUnlock()
	if !canUpload {
		return Entry{}, ErrPermissionDenied
	}
	resolved, err := m.resolveDirectory(rel)
	if err != nil {
		return Entry{}, err
	}
	name, err = cleanFilename(name)
	if err != nil {
		return Entry{}, err
	}
	createdName, err := createUniqueDir(resolved.RealPath, name, 0755)
	if err != nil {
		return Entry{}, err
	}
	dst := filepath.Join(resolved.RealPath, createdName)
	info, err := os.Stat(dst)
	if err != nil {
		return Entry{}, err
	}
	relPath := path.Join(resolved.CleanRel, info.Name())
	return Entry{Name: info.Name(), Path: slashRel(relPath), IsDir: true, ModTime: info.ModTime()}, nil
}

func (m *Manager) isWithinCurrentRoot(real string) bool {
	m.mu.RLock()
	rootReal := m.rootReal
	m.mu.RUnlock()
	return rootReal != "" && withinRoot(rootReal, real)
}

func trashName(rel string) string {
	prefix := time.Now().Format("20060102-150405.000000000") + "_"
	name := sanitizeFilename(prefix + strings.ReplaceAll(rel, "/", "_"))
	if filenameTooLong(name) {
		name = trimFilename(name, filenameLimit())
		if name == "" || name == prefix[:len(prefix)-1] {
			return strings.TrimSuffix(prefix, "_")
		}
	}
	return name
}
