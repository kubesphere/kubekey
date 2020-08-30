package main

import versionutil "k8s.io/apimachinery/pkg/util/version"

func main() {
	targetVerison := versionutil.MustParseSemantic("v1.18.6")
	println(targetVerison.Minor())
	println(targetVerison.Major())
}
