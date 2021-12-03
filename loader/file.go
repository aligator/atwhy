package loader

import (
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/aligator/nogo"
	"github.com/spf13/afero"
	"gitlab.com/tiffinger-thiel/crazydoc/tag"
)

type TagFinder interface {
	Find(filename string, reader io.Reader) (tags []tag.Raw, err error)
}

type File struct {
	FS             afero.Fs
	FileExtensions []string
}

// @DOC readme.ignore
// If you want to ignore files, just add a `.crazydocignore` to the root of your project.
// It follows the syntax of a `.gitignore` and you may also add `.crazydocignore` files to subfolders.

func (fl File) Load(dir string, finder TagFinder) ([]tag.Raw, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	filesystem := afero.NewBasePathFs(fl.FS, abs)
	allTags := make([]tag.Raw, 0)

	n := nogo.New(nogo.DotGitRule)

	err = afero.Walk(filesystem, ".", func(path string, info fs.FileInfo, err error) error {
		if ok, err := n.WalkFunc(afero.NewIOFS(filesystem), ".crazydocignore", path, info.IsDir(), err); !ok {
			return err
		}

		if info.IsDir() {
			return nil
		}

		hasFoundExtension := false

		for _, e := range fl.FileExtensions {
			fileName := info.Name()
			if strings.HasSuffix(fileName, e) {
				hasFoundExtension = true
			}
		}

		if !hasFoundExtension {
			return nil
		}

		file, err := filesystem.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		tags, err := finder.Find(path, file)
		if err != nil {
			return err
		}
		allTags = append(allTags, tags...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return allTags, nil
}