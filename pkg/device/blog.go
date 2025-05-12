//go:build !tinygo

package device

import (
	"encoding/json"
	"io/fs"
	"sort"
	"strings"
	"time"
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

	// Sort blogs by Date in reverse chronological order
	sort.Slice(blogs, func(i, j int) bool {
		// Parse dates in format "Nov. 10, 2024"
		// First remove the period after the month abbreviation
		datei := strings.Replace(blogs[i].Date, ".", "", 1)
		datej := strings.Replace(blogs[j].Date, ".", "", 1)

		ti, err := time.Parse("Jan 2, 2006", datei)
		if err != nil {
			panic(err.Error())
		}
		tj, err := time.Parse("Jan 2, 2006", datej)
		if err != nil {
			panic(err.Error())
		}
		return ti.After(tj) // Sort in reverse chronological order
	})

	return blogs
}
