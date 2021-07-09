package main

import (
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/gkhit/gscltmsd/service"
)

func main() {
	var (
		filename       string
		extension      string
		configfilename string
		dir            string
		err            error
		configpath     string
		opt            *service.Options
	)

	flag.StringVar(&configpath, "c", "", "full path to configuration json `file`")
	flag.Parse()

	filename = filepath.Base(os.Args[0])
	extension = filepath.Ext(filename)
	filename = filename[0 : len(filename)-len(extension)]

	if len(configpath) <= 0 {
		configfilename = filename + ".json"
		dir, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatalf("[ERROR] %v", err)
		}
		configpath = path.Join(dir, configfilename)
	}
	// configpath = "D:\\projects\\gscltmsd\\example.gscltmsd.json"
	// configpath = "/home/thinker/projects/gscltmsd/example.gscltmsd.json"
	opt = service.NewOptions()
	opt.FileLog.Filename = filename + ".log"
	err = opt.Load(configpath)
	if err != nil {
		log.Fatalf("[ERROR] Can't load configuration file. %v", err)
	}

	svc := service.New(opt)
	svc.Start()
}
