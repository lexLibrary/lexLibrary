package files

import (
	"bytes"
	"compress/gzip"
	"io"
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
	defer file.Close()

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
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, errors.Errorf("Asset %s is a Directory, not a File", name)
	}

	if _, ok := file.(httpgzip.NotWorthGzipCompressing); ok {
		return nil, errors.Errorf("Asset %s is not worth compressing", name)
	}

	if gz, ok := file.(httpgzip.GzipByter); ok {
		return gz.GzipBytes(), nil
	}

	// this shouldn't be necessary unless running in dev mode
	var b bytes.Buffer
	writer := gzip.NewWriter(&b)
	_, err = io.Copy(writer, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// AssetDir returns a file / dir listing for embedded assets
func AssetDir(name string) ([]string, error) {
	file, err := Assets.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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
