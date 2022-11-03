// boltdbweb is a webserver base GUI for interacting with BoltDB databases.
//
// For authorship see https://github.com/evnix/boltdbweb
// MIT license is included in repository
package main

import (
	"embed"
	"flag"
	"fmt"
	boltbrowserweb "github.com/evnix/boltdbweb/web"
	"net/http"
	"os"
	"path"
	"time"
)

import (
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

const version = "v1.0.0"

//go:embed web
var webAssets embed.FS

var (
	showHelp   bool
	db         *bolt.DB
	dbName     string
	port       string
	staticPath string
)

func usage(appName, version string) {
	fmt.Printf("Usage: %s [OPTIONS] [DB_NAME]", appName)
	fmt.Printf("\nOPTIONS:\n\n")
	flag.VisitAll(func(f *flag.Flag) {
		if len(f.Name) > 1 {
			fmt.Printf("    -%s, -%s\t%s\n", f.Name[0:1], f.Name, f.Usage)
		}
	})
	fmt.Printf("\n\nVersion %s\n", version)
}

func init() {
	// Read the static path from the environment if set.
	dbName = os.Getenv("BOLTDBWEB_DB_NAME")
	port = os.Getenv("BOLTDBWEB_PORT")
	staticPath = os.Getenv("BOLTDBWEB_STATIC_PATH")
	if port == "" {
		port = "8080"
	}
	// Setup for command line processing
	flag.BoolVar(&showHelp, "h", false, "display help")
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.StringVar(&dbName, "d", dbName, "Path of the database")
	flag.StringVar(&port, "p", port, "Port for the web-ui")
	flag.StringVar(&port, "port", port, "Port for the web-ui")
	flag.StringVar(&staticPath, "s", staticPath, "Path for the static content")
}

func main() {
	appName := path.Base(os.Args[0])
	flag.Parse()
	args := flag.Args()

	if showHelp == true {
		usage(appName, version)
		os.Exit(0)
	}

	// If non-flag options are included assume bolt db is specified.
	if len(args) > 0 {
		dbName = args[0]
	}

	if dbName == "" {
		usage(appName, version)
		log.Error("ERROR: Missing boltdb name")
		os.Exit(1)
	}

	fmt.Print(" ")
	log.Info("starting boltdb-browser..")

	var err error
	db, err = bolt.Open(dbName, 0600, &bolt.Options{Timeout: 2 * time.Second})
	boltbrowserweb.Db = db

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// OK, we should be ready to define/run web server safely.
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", boltbrowserweb.Index)

	r.GET("/buckets", boltbrowserweb.Buckets)
	r.POST("/createBucket", boltbrowserweb.CreateBucket)
	r.POST("/put", boltbrowserweb.Put)
	r.POST("/get", boltbrowserweb.Get)
	r.POST("/deleteKey", boltbrowserweb.DeleteKey)
	r.POST("/deleteBucket", boltbrowserweb.DeleteBucket)
	r.POST("/prefixScan", boltbrowserweb.PrefixScan)

	if staticPath != "" {
		log.Infof("using static path: %s", staticPath)
		r.Static("/static", staticPath)
	} else {
		log.Infof("using embedded static content")
		r.StaticFS("/static", http.FS(webAssets))
	}

	r.Run(":" + port)
}
