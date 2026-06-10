package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// 0020 adds a paid bool field to league_members so league admins can track
// which players have paid their entry fee.
func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("league_members")
		if err != nil {
			return err
		}
		if col.Fields.GetByName("paid") == nil {
			col.Fields.Add(&core.BoolField{Name: "paid"})
			if err := app.Save(col); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("league_members")
		if err != nil {
			return err
		}
		if col.Fields.GetByName("paid") != nil {
			col.Fields.RemoveByName("paid")
			return app.Save(col)
		}
		return nil
	})
}
