package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// 0021 adds a forecastUnlocked bool field to league_members so league admins
// can temporarily re-open an individual member's Forecast after the
// tournament has locked it (e.g. a player who never finished their bracket).
func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("league_members")
		if err != nil {
			return err
		}
		if col.Fields.GetByName("forecastUnlocked") == nil {
			col.Fields.Add(&core.BoolField{Name: "forecastUnlocked"})
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
		if col.Fields.GetByName("forecastUnlocked") != nil {
			col.Fields.RemoveByName("forecastUnlocked")
			return app.Save(col)
		}
		return nil
	})
}
