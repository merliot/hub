//go:build !tinygo

package device

import (
	"encoding/json"
	"io/fs"
	"sort"
)

// Each blog has a dir with two files, e.g.:
//
// blog
// ├── 2024-10-16-first-blog
// │   ├── blog.html
// │   └── blog.json
//
// The blog.json file is, e.g.:
//
//     {
//         "Title": "First Blog",
//         "Date": "Oct. 16, 2024"
//     }
//

type blog struct {
	Dir   string
	Title string
	Date  string
}

func (s *server) blogs() []blog {
	dirs, _ := fs.ReadDir(s.root.layeredFS, "blog")

	// Sort dirs in reverse order
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name() > dirs[j].Name()
	})

	blogs := make([]blog, 0, len(dirs))
	for _, dir := range dirs {
		var b = blog{Dir: dir.Name()}
		bytes, err := s.root.layeredFS.readFile("blog/" + dir.Name() + "/blog.json")
		if err != nil {
			panic(err.Error())
		}
		err = json.Unmarshal(bytes, &b)
		if err != nil {
			panic(err.Error())
		}
		blogs = append(blogs, b)
	}
	return blogs
}
