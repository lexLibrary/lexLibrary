//go:generate go run -tags=dev assets_generate.go
package files

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/shurcooL/httpgzip"
)

// Asset returns an embedded asset
func Asset(name string) ([]byte, error) {
	file, err := Assets.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, errors.Errorf("Asset %s is a Directory, not a File", name)
	}

	return ioutil.ReadAll(file)
}

// AssetCompressed returns a compressed embedded asset
func AssetCompressed(name string) ([]byte, error) {
	file, err := Assets.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, errors.Errorf("Asset %s is a Directory, not a File", name)
	}

	if _, ok := file.(httpgzip.NotWorthGzipCompressing); ok {
		return nil, errors.Errorf("Cannot get compressed Asset %s. It is Not Worth Compressing", name)
	}

	gz, ok := file.(httpgzip.GzipByter)
	if !ok {
		return nil, errors.Errorf("Cannot get compressed Asset %s", name)
	}

	return gz.GzipBytes(), nil
}

// AssetDir returns a file / dir listing for embedded assets
func AssetDir(name string) ([]string, error) {
	file, err := Assets.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.Errorf("Asset %s is a File, not a Directory", name)
	}

	ls, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}

	files := make([]string, len(ls))

	for i := range ls {
		files[i] = ls[i].Name()
	}
	return files, nil
}
