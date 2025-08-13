package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type rotatingWriter struct {
	path    string
	maxSize int64
	backups int
	cur     *os.File
	written int64
}

func newRotatingWriter(path string, maxSize int64, backups int) (io.WriteCloser, error) {
	if path == "" || maxSize <= 0 {
		return nopCloser{Writer: mustOpen(path)}, nil
	}
	w := &rotatingWriter{path: path, maxSize: maxSize, backups: backups}
	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}
func (w *rotatingWriter) open() error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.cur = f
	if fi, _ := f.Stat(); fi != nil {
		w.written = fi.Size()
	}
	return nil
}
func (w *rotatingWriter) rotate() error {
	if w.cur != nil {
		_ = w.cur.Close()
		w.cur = nil
	}
	for i := w.backups - 1; i >= 1; i-- {
		_ = os.Rename(fmt.Sprintf("%s.%d", w.path, i), fmt.Sprintf("%s.%d", w.path, i+1))
	}
	ts := time.Now().Format("20060102-150405")
	_ = os.Rename(w.path, fmt.Sprintf("%s.%s.1", w.path, ts))
	if w.backups > 0 {
		_ = os.Remove(fmt.Sprintf("%s.%d", w.path, w.backups+1))
	}
	w.written = 0
	return w.open()
}
func (w *rotatingWriter) Write(p []byte) (int, error) {
	if w.cur == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}
	if w.maxSize > 0 && w.written+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.cur.Write(p)
	w.written += int64(n)
	return n, err
}
func (w *rotatingWriter) Close() error {
	if w.cur != nil {
		return w.cur.Close()
	}
	return nil
}

type nopCloser struct{ io.Writer }

func (n nopCloser) Close() error { return nil }

func mustOpen(path string) *os.File {
	if path == "" {
		return os.Stdout
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	return f
}
