// package main is a module for recording audio and collecting to Viam
package main

import (
	"context"

	"github.com/edaniels/golog"
	"github.com/bee-ring/mic-record-viam-module/micrecord"
	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"
)

func main() {
	utils.ContextualMain(mainWithArgs, golog.NewDevelopmentLogger("micRecModule"))
}

func mainWithArgs(ctx context.Context, args []string, logger golog.Logger) error {
	micrecordModule, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	micrecordModule.AddModelFromRegistry(ctx, generic.API, micrecord.Model)

	err = micrecordModule.Start(ctx)
	defer micrecordModule.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
