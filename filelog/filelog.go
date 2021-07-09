package filelog

import (
	"log"
	"os"
	"path"

	lj "gopkg.in/natefinch/lumberjack.v2"
)

type (
	Options struct {
		Enable bool `json:"enable,omitempty"`
		// Directory to log to to when filelogging is enabled
		Directory string `json:"directory"`
		// Filename is the name of the logfile which will be placed inside the directory
		Filename string `json:"filename"`
		// MaxSize the max size in MB of the logfile before it's rolled
		MaxSize int `json:"max_size"`
		// MaxBackups the max number of rolled files to keep
		MaxBackups int `json:"max_backups"`
		// MaxAge the max age in days to keep a logfile
		MaxAge int `json:"max_age"`
	}
)

func NewWithOptions(o *Options) {

	if !o.Enable {
		return
	}

	if err := os.MkdirAll(o.Directory, 0744); err != nil {
		log.Fatalf("[ERROR] Can't create log directory: \"%s\". %v", o.Directory, err)
	}

	log.SetOutput(&lj.Logger{
		Filename:   path.Join(o.Directory, o.Filename),
		MaxBackups: o.MaxBackups, // files
		MaxSize:    o.MaxSize,    // megabytes
		MaxAge:     o.MaxAge,     // days
	})
}
