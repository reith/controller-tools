package deepcopy

import (
	"fmt"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	"sigs.k8s.io/controller-tools/pkg/loader"
)

// ImportsList keeps track of required imports, automatically assigning aliases
// to import statement.
type ImportsList struct {
	byPath  map[string]string
	byAlias map[string]string

	pkg *loader.Package
}

func NewImportsList(pkg *loader.Package) *ImportsList {
	return &ImportsList{
		byPath: make(map[string]string),
		// avoid confusing aliases by "reserving" the root package's name as an alias
		byAlias: map[string]string{pkg.Name: ""},
		pkg:     pkg,
	}
}

// NeedImport marks that the given package is needed in the list of imports,
// returning the ident (import alias) that should be used to reference the package.
func (l *ImportsList) NeedImport(importPath string) string {
	// we get an actual path from Package, which might include venddored
	// packages if running on a package in vendor.
	if ind := strings.LastIndex(importPath, "/vendor/"); ind != -1 {
		importPath = importPath[ind+8: /* len("/vendor/") */]
	}

	// check to see if we've already assigned an alias, and just return that.
	alias, exists := l.byPath[importPath]
	if exists {
		return alias
	}

	// otherwise, calculate an import alias by joining path parts till we get something unique
	restPath, nextWord := path.Split(importPath)

	for otherPath, exists := "", true; exists && otherPath != importPath; otherPath, exists = l.byAlias[alias] {
		if restPath == "" {
			// do something else to disambiguate if we're run out of parts and
			// still have duplicates, somehow
			alias += "x"
		}

		// can't have a first digit, per Go identifier rules, so just skip them
		for firstRune, runeLen := utf8.DecodeRuneInString(nextWord); unicode.IsDigit(firstRune); firstRune, runeLen = utf8.DecodeRuneInString(nextWord) {
			nextWord = nextWord[runeLen:]
		}

		// make a valid identifier by replacing "bad" characters with underscores
		nextWord = strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				return r
			}
			return '_'
		}, nextWord)

		alias = nextWord + alias
		if len(restPath) > 0 {
			restPath, nextWord = path.Split(restPath[:len(restPath)-1] /* chop off final slash */)
		}
	}

	l.byPath[importPath] = alias
	l.byAlias[alias] = importPath
	return alias
}

// ImportSpecs returns a string form of each import spec
// (i.e. `alias "path/to/import").  Aliases are only present
// when they don't match the package name.
func (l *ImportsList) ImportSpecs() []string {
	res := make([]string, 0, len(l.byPath))
	for importPath, alias := range l.byPath {
		pkg := l.pkg.Imports()[importPath]
		if pkg != nil && pkg.Name == alias {
			// don't print if alias is the same as package name
			// (we've already taken care of duplicates).
			res = append(res, fmt.Sprintf("%q", importPath))
		} else {
			res = append(res, fmt.Sprintf("%s %q", alias, importPath))
		}
	}
	return res
}
