package instance

import (
	service "github.com/mbh/himnario-back-end-go/internal/core/service"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/controller"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/database"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"
	"github.com/mbh/himnario-back-end-go/internal/infra/secondary/hymnary"
	"github.com/mbh/himnario-back-end-go/internal/infra/secondary/users"

	"github.com/mbh/himnario-back-end-go/internal/infra/primary/controllers/hymns"
	"github.com/mbh/himnario-back-end-go/internal/infra/primary/controllers/user"
	"github.com/mbh/libraries/golang/lightms"
)

func GetControllerHymns() lightms.PrimaryProcess {
	database.RunMigrations(property.GetDatabaseProperty().DatabaseSettings.Himnario)

	return controller.GetControllerInstance(
		[]controller.ControllerRunnable{
			user.NewUserController(
				service.NewUserService(
					users.NewClient(),
				),
			),
			hymns.NewHymnController(
				service.NewHymnsService(
					hymnary.NewHymns(),
				),
			),
		},
	)
}
