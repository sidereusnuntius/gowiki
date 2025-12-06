package diff

import "github.com/sergi/go-diff/diffmatchpatch"

var dmp *diffmatchpatch.DiffMatchPatch

func init() {
	dmp = diffmatchpatch.New()
}

func FindPatches(text1, text2 string) string {
	diffs := dmp.DiffMain(text1, text2, false)
	return dmp.PatchToText(dmp.PatchMake(diffs))
}