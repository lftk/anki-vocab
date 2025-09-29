package dict

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Queryer interface {
	Query(ctx context.Context, word string) ([]byte, error)
}

type Pronouncer interface {
	Pronounce(ctx context.Context, word, accent, format string) (io.ReadCloser, error)
}

type QueryCapabilities struct {
	AI bool
}

type PronounceCapabilities struct {
	Accents []string // 支持的口音
	Formats []string // 支持的音频格式
}

type Capabilities struct {
	Query     *QueryCapabilities
	Pronounce *PronounceCapabilities
}

type Dict struct {
	Queryer      Queryer
	Pronouncer   Pronouncer
	Capabilities *Capabilities
}

type cachedQueryer struct {
	dir string
	Queryer
}

func CachedQueryer(dir string, q Queryer) Queryer {
	if cq, ok := q.(*cachedQueryer); ok {
		if cq.dir == dir {
			return q
		}
	}
	return &cachedQueryer{
		dir:     dir,
		Queryer: q,
	}
}

func (q *cachedQueryer) Query(ctx context.Context, word string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	path := filepath.Join(q.dir, fmt.Sprintf("%s.json", word))
	b, err := os.ReadFile(path)
	switch {
	case err == nil:
		return b, nil
	case !errors.Is(err, fs.ErrNotExist):
		return nil, err
	default:
		b, err = q.Queryer.Query(ctx, word)
		if err != nil {
			return nil, err
		}
		return b, os.WriteFile(path, b, 0644)
	}
}

type cachedPronouncer struct {
	dir string
	Pronouncer
}

func CachedPronouncer(dir string, p Pronouncer) Pronouncer {
	if cp, ok := p.(*cachedPronouncer); ok {
		if cp.dir == dir {
			return p
		}
	}
	return &cachedPronouncer{
		dir:        dir,
		Pronouncer: p,
	}
}

func (cp *cachedPronouncer) Pronounce(ctx context.Context, word, accent, format string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	path := filepath.Join(cp.dir, fmt.Sprintf("%s_%s.%s", word, accent, format))
	f, err := os.Open(path)
	switch {
	case err == nil:
		return f, nil
	case !errors.Is(err, fs.ErrNotExist):
		return nil, err
	default:
		audio, err := cp.Pronouncer.Pronounce(ctx, word, accent, format)
		if err != nil {
			return nil, err
		}
		w, err := os.Create(path)
		if err != nil {
			_ = audio.Close()
			return nil, err
		}
		return &teeReadCloser{r: audio, w: w}, nil
	}
}

type teeReadCloser struct {
	r io.ReadCloser
	w io.WriteCloser
}

func (t *teeReadCloser) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

func (t *teeReadCloser) Close() error {
	return errors.Join(t.r.Close(), t.w.Close())
}
