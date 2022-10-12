package plugindef

import (
	"embed"
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/grafana/grafana/pkg/cuectx"
	"github.com/grafana/thema"
)

var (
	lineageOnce sync.Once
	clin        thema.ConvergentLineage[*Plugindef]
	clinerr     error
)

//go:embed plugindef.cue
var cueFS embed.FS
var themaFS fs.FS

func init() {
	var err error

	themaFS, err = cuectx.PrefixWithGrafanaCUE(filepath.Join("pkg", "plugins", "plugindef"), cueFS)
	if err != nil {
		panic(err)
	}
	_, err = Lineage(cuectx.GrafanaThemaRuntime())
	if err != nil {
		panic(err)
	}
}

// Lineage returns a [thema.ConvergentLineage] for the 'plugindef' lineage.
//
// The lineage is the canonical specification of plugindef. It contains
// all schema versions that have ever existed for plugindef,
// and the lenses that allow valid instances of one schema in the lineage to
// be translated to another schema in the lineage.
//
// This function will return an error if the [Thema invariants] are not met by
// the lineage defined in plugindef.cue.
//
// [Thema's general invariants]: https://github.com/grafana/thema/blob/main/docs/invariants.md
func Lineage(rt *thema.Runtime, opts ...thema.BindOption) (thema.ConvergentLineage[*Plugindef], error) {
	allrt := cuectx.GrafanaThemaRuntime()
	if rt == nil || rt == allrt {
		lineageOnce.Do(func() {
			clin, clinerr = doLineage(allrt, opts...)
		})
		return clin, clinerr
	}

	return doLineage(rt, opts...)
}
