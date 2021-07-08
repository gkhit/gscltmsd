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

	if len(configpath) <= 0 {
		filename = filepath.Base(os.Args[0])
		extension = filepath.Ext(filename)
		configfilename = filename[0 : len(filename)-len(extension)]
		configfilename += ".json"
		dir, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		configpath = path.Join(dir, configfilename)
	}
	// configpath = "D:\\projects\\gscltmsd\\example.gscltmsd.json"
	// configpath = "/home/thinker/projects/gscltmsd/example.gscltmsd.json"
	opt = service.NewOptions()
	err = opt.Load(configpath)
	if err != nil {
		log.Fatalf("[ERROR] Can't load configuration file. %v", err)
	}

	svc := service.New(opt)
	svc.Start()
}
