package migrations

import (
	"github.com/nyaruka/goflow/utils/uuids"

	"github.com/Masterminds/semver"
)

func init() {
	registerMigration(semver.MustParse("13.1"), AsParsed(Migrate13_1))
}

// Migrate13_1 adds UUID to send_msg templating
func Migrate13_1(f Flow) Flow {
	for _, node := range f.Nodes() {
		for _, action := range node.Actions() {
			if action.Type() == "send_msg" {
				templating, _ := action.Def["templating"].(map[string]interface{})
				if templating != nil {
					templating["uuid"] = uuids.New()
				}
			}
		}
	}
	return f
}
