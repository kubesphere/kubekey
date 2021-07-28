package render

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	regDoublePat = regexp.MustCompile(`\{\{.*?\}\}`)
	regRep       = regexp.MustCompile(`[\{\} ]`)
)

func Render(input string, replaceVars map[string]string) (string, error) {
	res := input
	oldVars := regDoublePat.FindAllString(input, -1)
	for _, v := range oldVars {
		// delete the " ", "{" and "}"
		realVar := regRep.ReplaceAllString(v, "")
		if shoutVar, ok := replaceVars[realVar]; !ok || regDoublePat.MatchString(shoutVar) {
			return "", fmt.Errorf("render failed: [%s]", realVar)
		} else {
			res = strings.ReplaceAll(res, v, shoutVar)
		}
	}
	return res, nil
}
