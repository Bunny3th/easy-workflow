package config

type Config_App struct {
	LogPath string
	UploadPath string
}

var App Config_App

func(i *Initializer) App(){
	App.LogPath = cfg.GetString("App.LogPath")
	App.UploadPath=cfg.GetString("App.UploadPath")
}
