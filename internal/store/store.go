package store

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alesgaroth/tasksd/internal/core"
)

type Store struct {
	repoDir string
}

func New(repoDir string) *Store {
	return &Store{repoDir: repoDir}
}

func (s *Store) ListItems() ([]core.Item, error) {
	var items []core.Item
	dirs := []string{
		filepath.Join(s.repoDir, "items"),
		filepath.Join(s.repoDir, "tickets"),
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if !info.IsDir() {
			continue
		}

		err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".md") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read file %s: %w", path, err)
			}

			item, err := core.ParseItem(string(content))
			if err != nil {
				return fmt.Errorf("parse file %s: %w", path, err)
			}

			items = append(items, item)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

func ValidatePath(targetPath string) ([]string, int, error) {
	info, err := os.Stat(targetPath)
	if err != nil {
		return nil, 0, err
	}

	var validationErrors []string
	count := 0

	if !info.IsDir() {
		if !strings.HasSuffix(targetPath, ".md") {
			return nil, 0, nil
		}
		count++
		content, err := os.ReadFile(targetPath)
		if err != nil {
			return nil, 0, err
		}
		if _, err := core.ParseItem(string(content)); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("%s: %v", targetPath, err))
		}
		return validationErrors, count, nil
	}

	dirsToWalk := []string{
		filepath.Join(targetPath, "items"),
		filepath.Join(targetPath, "tickets"),
	}

	hasSubdir := false
	for _, d := range dirsToWalk {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			hasSubdir = true
		}
	}
	if !hasSubdir {
		dirsToWalk = []string{targetPath}
	}

	for _, dir := range dirsToWalk {
		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			continue
		}
		err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".md") {
				return nil
			}

			count++
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := core.ParseItem(string(content)); err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("%s: %v", path, err))
			}
			return nil
		})
		if err != nil {
			return nil, 0, err
		}
	}

	return validationErrors, count, nil
}
