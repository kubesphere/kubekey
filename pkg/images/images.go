package images

import (
	"fmt"
)

type Image struct {
	Prefix string
	Repo   string
	Tag    string
}

func (image *Image) NewImage() string {
	return fmt.Sprintf("%s%s/%s", image.Prefix, image.Repo, image.Tag)
}

func GetImagePrefix(privateRegistry, ns string) string {
	var prefix string
	if privateRegistry == "" {
		if ns == "" {
			prefix = ""
		} else {
			prefix = fmt.Sprintf("%s/", ns)
		}
	} else {
		if ns == "" {
			prefix = fmt.Sprintf("%s/library/", privateRegistry)
		} else {
			prefix = fmt.Sprintf("%s/%s/", privateRegistry, ns)
		}
	}
	return prefix
}
