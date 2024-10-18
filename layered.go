//go:build !tinygo

package hub

import (
	"html/template"
	"io/fs"
	"io/ioutil"
)

// layeredFS is a layered fs, built up from individual file systems
type layeredFS struct {
	layers []fs.ReadFileFS
}

// Stack adds fs to the layered fs.  Order matters: first added is lowest
// in priority when searching for a file name in the layered fs.
func (lfs *layeredFS) stack(fsys fs.ReadFileFS) {
	lfs.layers = append(lfs.layers, fsys)
}

func (lfs layeredFS) newest(name string) (fs.File, error) {

	// Start with newest (last added) FS, giving newer FSes priority over
	// older FSes when searching for file name.  The first FS with a
	// matching file name wins.

	for i := len(lfs.layers) - 1; i >= 0; i-- {
		fsys := lfs.layers[i]
		if file, err := fsys.Open(name); err == nil {
			return file, nil
		}
	}

	return nil, fs.ErrNotExist
}

// Open a file by name
func (lfs layeredFS) Open(name string) (fs.File, error) {
	return lfs.newest(name)
}

// readFile reads a file as []byte
func (lfs layeredFS) readFile(name string) ([]byte, error) {
	file, err := lfs.newest(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

// parseFS returns a template by parsing the layered file system for the
// template name matching the pattern name
func (lfs layeredFS) parseFS(pattern string, funcMap template.FuncMap) (*template.Template, error) {

	// Iterate from oldest (first added) FS to newest FS, building a "main"
	// template with pattern matching templates from each FS.  The winner
	// for when templates have the same name will be the last one added to
	// the main template (newest FS wins).

	mainTmpl := template.New("main").Funcs(funcMap)

	for _, fsys := range lfs.layers {
		_, err := mainTmpl.ParseFS(fsys, pattern)
		if err != nil {
			return nil, err
		}
	}

	return mainTmpl, nil
}
