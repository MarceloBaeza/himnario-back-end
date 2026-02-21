package main

import (
	"github.com/mbh/libraries/golang/lightms"

	"github.com/mbh/himnario-back-end-go/internal/infra/config/instance"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"
)

func main() {
	registerProperties()
	registerPrimaries()
	lightms.Run()
}

func registerPrimaries() {
	lightms.AddPrimary(instance.GetControllerHymns)
}

func registerProperties() {
	lightms.SetPropFilePath("./resources/properties.yml")
	lightms.AddProperty(property.GetServerProperty())
	lightms.AddProperty(property.GetDatabaseProperty())
	lightms.AddProperty(property.GetJwtProperty())
	lightms.AddProperty(property.GetValidationsProperty())
}
