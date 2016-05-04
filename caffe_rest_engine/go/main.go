package main

// #include <stdlib.h>
import "C"

//import "unsafe"
import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

var (
	debugMode       bool
	debugLog        string
	errorLog        string
	noPreloadModels bool

	debugLogger *log.Logger
	errorLogger *log.Logger

	logFlags = log.Lshortfile | log.Ltime | log.Lmicroseconds
)

func preload() {
	var modelGob []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS_BUCKET)
		c := b.Cursor()
		var model Model
		for k, v := c.First(); k != nil; k, v = c.Next() {
			modelGob = make([]byte, len(v))
			copy(modelGob, v)
			buf := bytes.NewBuffer(modelGob)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&model)
			if err != nil {
				continue
			}
			LRU.PushBack(model.Name)
			InitializeModel(&model)
		}
		return nil
	})

	if err != nil {
		panic("error in transaction! " + err.Error())
	}
}
func setupLoggers() {
	var debugFile io.Writer
	var err error
	if debugLog == "" {
		debugFile = os.Stdout
	} else {
		debugFile, err = os.OpenFile(debugLog, os.O_WRONLY, 0644)
		if err != nil {
			panic("Failed to open debug log: " + err.Error())
		}
	}
	debugLogger = log.New(debugFile, "DEBUG: ", logFlags)

	var errorFile io.Writer
	if errorLog == "" {
		errorFile = os.Stderr
	} else {
		errorFile, err = os.OpenFile(errorLog, os.O_WRONLY, 0644)
		if err != nil {
			panic("Failed to open error log: " + err.Error())
		}
	}

	errorLogger = log.New(errorFile, "Error: ", logFlags)

}

func Debugf(format string, v ...interface{}) {
	if debugMode {
		debugLogger.Printf(format, v...)
	}
}

func LogTimef(operation string, start time.Time, v ...interface{}) {
	duration := (time.Now().UnixNano() - start.UnixNano()) / 1000000
	Debugf(fmt.Sprintf("%v took %vs (%vms)", operation, float64(duration)/1000.0, duration), v...)
}

var batch_daemon *BatchDaemon = NewBatchDaemon()

func main() {
	runtime.GOMAXPROCS(48)
	nworkers := flag.Int("n", 4, "Enter the number of workers wanted.")
	flag.StringVar(&errorLog, "errorLog",
		"", "File location for error log. defaults to stderr",
	)
	flag.StringVar(&debugLog, "debugLog", "",
		string("File location for debug log. ")+
			string("Only meaningful if -debug is set. Defaults to stdout. "),
	)

	flag.BoolVar(&debugMode, "debug", false,
		string("Enables debug mode, which has more")+
			string("verbose logging and times certain operations."))
	flag.BoolVar(&noPreloadModels, "noPreloadModels", false, "Turn off model preloading.")
	flag.Parse()

	setupLoggers()

	if noPreloadModels {
		fmt.Println("Skipping preload...")
	} else {
		fmt.Println("Preloading and pre-init'ing models")
		preload()
		fmt.Println("Finished prefetching models into CPU Ram")
	}

	fmt.Println("Starting the dispatcher!")
	fmt.Println("nworker", *nworkers)
	dis := NewDispatcher("placeholder", *nworkers)
	dis.StartDispatcher()
	fmt.Println("Starting Background Batching Daemon")
	batch_daemon.Start()
	fmt.Println("Registering HTTP Function")
	http.HandleFunc("/classify", JobHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/list", ListHandler)
	fmt.Println("HTTP Server listening on 127.0.0.1:8000")
	errhttp := http.ListenAndServe("0.0.0.0:8000", nil)
	if errhttp != nil {
		fmt.Println(errhttp.Error())
	}
}
