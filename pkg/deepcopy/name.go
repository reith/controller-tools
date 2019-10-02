package deepcopy

import (
	"fmt"
	"go/types"

	"sigs.k8s.io/controller-tools/pkg/loader"
)

// NamingInfo holds package and syntax for referencing a field, type,
// etc.  It's used to allow lazily marking import usage.
// You should generally retrieve the syntax using Syntax.
type NamingInfo struct {
	// typeInfo is the type being named.
	typeInfo     types.Type
	nameOverride string
}

// Syntax calculates the code representation of the given type or name,
// and marks that is used (potentially marking an import as used).
func (n *NamingInfo) Syntax(basePkg *loader.Package, imports *ImportsList) string {
	if n.nameOverride != "" {
		return n.nameOverride
	}

	// NB(directxman12): typeInfo.String gets us most of the way there,
	// but fails (for us) on named imports, since it uses the full package path.
	switch typeInfo := n.typeInfo.(type) {
	case *types.Named:
		// register that we need an import for this type,
		// so we can get the appropriate alias to use.
		typeName := typeInfo.Obj()
		otherPkg := typeName.Pkg()
		if otherPkg == basePkg.Types {
			// local import
			return typeName.Name()
		}
		alias := imports.NeedImport(loader.NonVendorPath(otherPkg.Path()))
		return alias + "." + typeName.Name()
	case *types.Basic:
		return typeInfo.String()
	case *types.Pointer:
		return "*" + (&NamingInfo{typeInfo: typeInfo.Elem()}).Syntax(basePkg, imports)
	case *types.Slice:
		return "[]" + (&NamingInfo{typeInfo: typeInfo.Elem()}).Syntax(basePkg, imports)
	case *types.Map:
		return fmt.Sprintf(
			"map[%s]%s",
			(&NamingInfo{typeInfo: typeInfo.Key()}).Syntax(basePkg, imports),
			(&NamingInfo{typeInfo: typeInfo.Elem()}).Syntax(basePkg, imports))
	default:
		basePkg.AddError(fmt.Errorf("name requested for invalid type %s", typeInfo))
		return typeInfo.String()
	}
}
