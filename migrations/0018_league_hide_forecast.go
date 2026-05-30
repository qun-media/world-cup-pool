package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// 0018 adds a hideForecast bool field to leagues so admins can prevent
// members from viewing each other's pre-tournament bracket predictions.
func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("leagues")
		if err != nil {
			return err
		}
		if col.Fields.GetByName("hideForecast") == nil {
			col.Fields.Add(&core.BoolField{Name: "hideForecast"})
			if err := app.Save(col); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("leagues")
		if err != nil {
			return err
		}
		if col.Fields.GetByName("hideForecast") != nil {
			col.Fields.RemoveByName("hideForecast")
			return app.Save(col)
		}
		return nil
	})
}
