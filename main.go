package main

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/gkhit/gscltmsd/service"

	_ "github.com/denisenkom/go-mssqldb"
)

func main() {
	var (
		filename       string
		extension      string
		configfilename string
		dir            string
		err            error
		configpath     string
		opt            *service.ServiceOptions
		svc            *service.Service
	)
	filename = filepath.Base(os.Args[0])
	extension = filepath.Ext(filename)
	configfilename = filename[0 : len(filename)-len(extension)]
	configfilename += ".json"
	dir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	configpath = path.Join(dir, configfilename)
	// configpath = "/home/thinker/projects/go/src/gkhit.ru/scada/mqcltmssvc/mqcltmssvc.json"
	opt = service.NewOptions()
	err = opt.Load(configpath)
	if err != nil {
		log.Fatal(err)
	}

	svc = service.NewService(opt)
	svc.Start()
}
