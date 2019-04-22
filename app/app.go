package app

import (
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/manager"
)

type App struct {
	config  config.Config
	manager manager.Manager
}

func NewApp(config config.Config) *App {
	return &App{
		config:  config,
		manager: *manager.NewManager(config),
	}
}

func (a *App) InitializeComponents(config config.Config) error {
	err := a.manager.LoadPlugins(config)
	if err != nil {
		config.Logger.Panic(err)
	}

	err = a.manager.InitPlugins(config)
	if err != nil {
		config.Logger.Panic(err)
	}

	return err
}

func (a *App) StartComponents(config config.Config) error {
	err := a.manager.StartPlugins(config)

	if err != nil {
		config.Logger.Panic(err)
	}

	return err
}

func (a *App) InitializePipelines(config config.Config) error {
	err := a.manager.InitPipelines(config)

	if err != nil {
		config.Logger.Panic(err)
	}

	return err
}

func (a *App) StartPipelines(config config.Config) error {
	err := a.manager.StartPipelines(config)

	if err != nil {
		config.Logger.Panic(err)
	}

	return err
}
