package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Disable PocketBase's built-in "new device" auth alert emails for both
// regular users and superusers. The alerts fire on every login from an
// unrecognised browser/IP, which produces unwanted email noise.
func init() {
	m.Register(func(app core.App) error {
		for _, name := range []string{"users", core.CollectionNameSuperusers} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			if !c.AuthAlert.Enabled {
				continue
			}
			c.AuthAlert.Enabled = false
			if err := app.Save(c); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		for _, name := range []string{"users", core.CollectionNameSuperusers} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			if c.AuthAlert.Enabled {
				continue
			}
			c.AuthAlert.Enabled = true
			if err := app.Save(c); err != nil {
				return err
			}
		}
		return nil
	})
}
