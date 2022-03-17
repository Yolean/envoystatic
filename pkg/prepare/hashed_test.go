package prepare_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yolean/envoystatic/v1/pkg/prepare"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestExistence(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()
	in := t.TempDir()
	out := t.TempDir()
	_, err := prepare.NewHashed(in, out)
	if err == nil {
		t.Errorf("Should fail on existent out")
	}
	if err = os.Remove(out); err != nil {
		t.Errorf("Failed to remove out %s: %v", out, err)
	}
	_, err = prepare.NewHashed(in, out)
	if err != nil {
		t.Errorf("Should accept existent in and nonexistent out, got %v", err)
	}
	if err = os.Remove(in); err != nil {
		t.Errorf("Failed to remove tmp dir %s: %v", in, err)
	}
	_, err = prepare.NewHashed(in, "/tmp/whatever")
	if err == nil {
		t.Errorf("Should fail if in-path does not exist")
	}
	if !strings.HasSuffix(err.Error(), "no such file or directory") {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestHtml01(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	// TODO to verify that we stick with the FS abstraction we could instead of tempdir try
	// https://www.gopherguides.com/articles/golang-1.16-io-fs-improve-test-performance
	// https://pkg.go.dev/github.com/psanford/memfs
	out := t.TempDir()
	os.Remove(out)
	p, err := prepare.NewHashed("../../tests/html01", out)
	if err != nil {
		t.Errorf("Failed to initialize pipeline: %s", err.Error())
	}

	content, err := p.Process()
	if err != nil {
		t.Errorf("Failed to process: %s", err.Error())
	}
	if content == nil {
		t.Errorf("content is nil")
		return
	}
	if content.Items == nil {
		t.Errorf("Items is nil")
		return
	}

	tests := []struct {
		relpath  string
		mimetype string
		size     int64
		content  string
	}{
		{"Dockerfile", "", 0, ""},
		{"html01.sh", "", 0, ""},
		{"index.html", "text/html; charset=utf-8", 0, ""},
		{"script.js", "application/javascript", 25, "63314764b32e0f86ebc1b32a734cba2dabc4945b7897fc024f37f0bf16ed4226"},
		{},
		{"subdir/text.txt", "text/plain; charset=utf-8", 0, ""},
	}

	if len(content.Items) != len(tests) {
		t.Errorf("expected item count %d but got %d", len(tests), len(content.Items))
	}

	for i, test := range tests {
		item := content.Items[i]
		if test.relpath != "" {
			if item.Path != test.relpath {
				t.Errorf("expected relpath %d to be %s but got %s", i, test.relpath, item.Path)
			}
		}
		if test.mimetype != "" {
			if item.ContentType != test.mimetype {
				t.Errorf("expected mimetype %d to be %s but got %s", i, test.mimetype, item.ContentType)
			}
		}
		if test.content != "" {
			if test.size == 0 {
				// makes it easier to spot file changes in fixtures
				t.Errorf("must expect size if expecting content, at index %d", i)
			}
			outfile, err := os.Stat(filepath.Join(out, test.content))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					t.Errorf("out file was not created: %s", test.relpath)
				} else {
					t.Errorf("failed to stat out file: %s", test.relpath)
				}
			}
			if test.size != 0 {
				if outfile.Size() != test.size {
					t.Errorf("expected size %d to be %d but got %d", i, test.size, outfile.Size())
				}
			}
		}
	}
}
