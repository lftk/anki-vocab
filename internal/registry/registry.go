package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/lftk/anki-vocab/internal/dict"
	"github.com/lftk/anki-vocab/internal/dict/volcengine"
	"github.com/lftk/anki-vocab/internal/dict/youdao"
)

type config struct {
	Youdao     youdao.Config     `yaml:"youdao"`
	Volcengine volcengine.Config `yaml:"volcengine"`
}

func loadConfig(path string) (*config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err = yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type Registry struct {
	dicts map[string]*dict.Dict
	cache string
	cfg   *config
}

func New(cfgPath, cacheDir string) (*Registry, error) {
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	return &Registry{
		dicts: make(map[string]*dict.Dict),
		cache: cacheDir,
		cfg:   cfg,
	}, nil
}

func (r *Registry) LoadOrNew(name string) (*dict.Dict, error) {
	if d, ok := r.dicts[name]; ok {
		return d, nil
	}
	d, err := r.New(name)
	if err == nil {
		r.dicts[name] = d
	}
	return d, err
}

func (r *Registry) New(name string) (*dict.Dict, error) {
	fn, ok := dicts[name]
	if !ok {
		return nil, fmt.Errorf("unknown dictionary: %q", name)
	}
	d := fn(r.cfg)
	if r.cache != "" {
		dir := filepath.Join(r.cache, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		if d.Queryer != nil {
			d.Queryer = dict.CachedQueryer(dir, d.Queryer)
		}
		if d.Pronouncer != nil {
			d.Pronouncer = dict.CachedPronouncer(dir, d.Pronouncer)
		}
	}
	return d, nil
}

var dicts = map[string]func(*config) *dict.Dict{
	"youdao": func(cfg *config) *dict.Dict {
		return youdao.New(&cfg.Youdao)
	},
	"volcengine": func(cfg *config) *dict.Dict {
		return volcengine.New(&cfg.Volcengine)
	},
}
