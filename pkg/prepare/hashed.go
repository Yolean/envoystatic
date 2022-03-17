package prepare

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"

	"github.com/yolean/envoystatic/v1/pkg/routeconfig"
	"go.uber.org/zap"
)

type Hashed struct {
	in      fs.FS
	out     fs.FS
	outFile func(path string, content []byte) error
}

// NewHashed should not be configurable, should neither do transforms nor add indexes
func NewHashed(in string, out string) (*Hashed, error) {
	p := &Hashed{}
	p.in = os.DirFS(in)
	p.out = os.DirFS(out)

	instat, err := fs.Stat(p.in, ".")
	if err != nil {
		zap.L().Error("In path stat failed", zap.String("path", in), zap.Error(err))
		return nil, err
	}
	if !instat.IsDir() {
		zap.L().Error("In path is not a dir", zap.String("path", in), zap.Error(err))
		return nil, errors.New("in path is not a dir")
	}
	_, err = fs.Stat(p.out, ".")
	if err == nil {
		zap.L().Error("Out path exists already", zap.String("path", out))
		return nil, errors.New("out path exists already")
	}
	err = os.Mkdir(out, 0755)
	if err != nil {
		zap.L().Error("Failed to create out path", zap.String("path", out))
		return nil, err
	}
	p.outFile = func(path string, content []byte) error {
		// is there a WriteFile for fs.FS?
		abs := filepath.Join(out, path)
		// note that this impl won't create directories
		return ioutil.WriteFile(abs, content, 0644)
	}

	return p, nil
}

func (p *Hashed) Process() (*routeconfig.RouteContent, error) {
	content := &routeconfig.RouteContent{
		Items: []*routeconfig.ResponseItem{},
	}
	err := fs.WalkDir(p.in, ".", func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			zap.L().Error("in walk failed", zap.String("path", path))
			return err
		}
		if f.IsDir() {
			return nil
		}
		zap.L().Debug("walked",
			zap.String("path", path),
			zap.String("name", f.Name()),
		)
		items, err := p.file(path, f)
		if err != nil {
			zap.L().Error("failed to process file", zap.String("path", path), zap.Error(err))
			return err
		}
		content.Items = append(content.Items, items...)
		return nil
	})
	if err != nil {
		zap.L().Error("in WalkDir failed", zap.Error(err))
	}
	return content, nil
}

func (p *Hashed) file(path string, f fs.DirEntry) ([]*routeconfig.ResponseItem, error) {
	item := &routeconfig.ResponseItem{
		Path: path,
	}
	info, err := f.Info()
	if err != nil {
		zap.L().Error("in file info failed", zap.String("path", path), zap.Error(err))
	}
	item.ContentLength = info.Size()
	ext := filepath.Ext(f.Name())
	item.ContentType = mime.TypeByExtension(ext)

	infile, err := fs.ReadFile(p.in, path)
	if err != nil {
		zap.L().Error("in file read failed", zap.String("path", path), zap.Error(err))
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(infile))
	item.ETag = fmt.Sprintf(`"%s"`, hash)
	item.ContentPath = hash
	err = p.outFile(item.ContentPath, infile)
	if err != nil {
		zap.L().Error("out file write failed",
			zap.String("in", path),
			zap.String("out", item.ContentPath),
			zap.Error(err),
		)
	}

	return []*routeconfig.ResponseItem{item}, nil
}
