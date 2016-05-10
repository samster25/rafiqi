package main

// #include <stdlib.h>
// #include <classification.h>
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
	"strconv"
	"time"
)

var (
	debugMode       bool
	debugLog        string
	errorLog        string
	maxGPUMemUsage  uint64
	QUANTA          int64
	MAX_BATCH_AMT   int
	NUM_CONTEXTS    int
	useSync         bool
	maxCachedModels int

	initialMemoryUsage uint64

	debugLogger *log.Logger
	errorLogger *log.Logger

	logFlags = log.Lshortfile | log.Ltime | log.Lmicroseconds
)

func preloadModel(model Model) error {
	/*	var isFirstLoad bool
		if LRU.Len() == 0 {
			isFirstLoad = true
		} else {
			isFirstLoad = false
		}
	*/
	LRU.PushBack(model.Name)
	batch_daemon.ModelInfo[model.Name] = NewModelEntry()

	//var beforeUsage uint64
	/*
		if model.ModelSize == 0 {
			//beforeUsage = MemoryManager.GetCurrentMemUsage()

			MemoryManager.LoadModel(model)

			if isFirstLoad {
				// Find out baseline usage
				modelUsage := MemoryManager.GetCurrentMemUsage()
				initialMemoryUsage = MemoryManager.GetStaticGPUUsage()
				model.ModelSize = modelUsage - initialMemoryUsage
			} else {
				model.ModelSize = MemoryManager.GetCurrentMemUsage() - beforeUsage
			}

			fmt.Println("About to update", model.Name, "to have size", model.ModelSize)
		} else {*/
	MemoryManager.LoadModel(model)
	//}

	/*

		var encModel bytes.Buffer

		enc := gob.NewEncoder(&encModel)
		err := enc.Encode(model)
		if err != nil {
			return err
		}

		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(MODELS_BUCKET)
			err = b.Put([]byte(model.Name), encModel.Bytes())
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return err
		}*/

	return nil
}

func preload() {
	var modelGob []byte

	models := make([]Model, 0)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS_BUCKET)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var model Model

			modelGob = make([]byte, len(v))
			copy(modelGob, v)
			buf := bytes.NewBuffer(modelGob)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&model)
			if err != nil {
				continue
			}
			models = append(models, model)
		}
		return nil
	})

	if err != nil {
		panic("error acquiring list of models to preload! " + err.Error())
	}

	for _, model := range models {
		err = preloadModel(model)
		if err != nil {
			panic("error preloading " + model.Name + ": " + err.Error())
		}
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

func DebugPanic(format string, v ...interface{}) {
	if debugMode {
		panic(fmt.Sprintf(format, v...))
	} else {
		errorLogger.Printf(format, v...)
	}
}

func LogTimef(operation string, start time.Time, v ...interface{}) {
	duration := (time.Now().UnixNano() - start.UnixNano()) / 1000000
	Debugf(fmt.Sprintf("%v took %vs (%vms)", operation, float64(duration)/1000.0, duration), v...)
}

var batch_daemon *BatchDaemon = NewBatchDaemon()

func ChangeParamsHandler(w http.ResponseWriter, r *http.Request) {
	quantaS := r.FormValue("quanta")
	batchSizeS := r.FormValue("batchSize")

	if quantaS == "" || batchSizeS == "" {
		w.Write([]byte("missing quanta or batch size"))
		return
	}

	quanta, _ := strconv.ParseInt(quantaS, 10, 64)
	batchSize, _ := strconv.Atoi(batchSizeS)

	QUANTA = quanta
	MAX_BATCH_AMT = batchSize
	w.Write([]byte("successfully changed"))

}

func main() {
	runtime.GOMAXPROCS(32)

	nworkers := flag.Int("n", 4, "Enter the number of workers wanted.")
	flag.StringVar(&errorLog, "errorLog",
		"", "File location for error log. defaults to stderr",
	)
	flag.StringVar(&debugLog, "debugLog", "",
		string("File location for debug log. ")+
			string("Only meaningful if -debug is set. Defaults to stdout. "),
	)

	flag.BoolVar(&useSync, "useSync", false, "Use synchronous gpu memcopy's")

	flag.BoolVar(&debugMode, "debug", false,
		string("Enables debug mode, which has more")+
			string("verbose logging and times certain operations."))

	flag.IntVar(&MAX_BATCH_AMT, "maxBatch", 64, "Maximum batch size")
	flag.IntVar(&NUM_CONTEXTS, "numContexts", 2, "Number of Caffe contexts/model")
	flag.IntVar(&maxCachedModels, "maxCachedModels", -1, "Max number of cached models. Use -1 to disable model counting.")

	totalGPUMem := int64(C.get_total_gpu_memory())

	flag.Uint64Var(&maxGPUMemUsage, "maxCacheSize", uint64(totalGPUMem), "Maximum amount of space used in GPU memory at one time (in bytes).")

	flag.Int64Var(&QUANTA, "quanta", 3, "Watchdog quanta")

	flag.Parse()

	setupLoggers()

	fmt.Println("Preloading and pre-init'ing models")
	preload()
	fmt.Println("Finished prefetching models into CPU Ram")

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
	http.HandleFunc("/change_params", ChangeParamsHandler)
	fmt.Println("HTTP Server listening on 127.0.0.1:8000")
	errhttp := http.ListenAndServe("0.0.0.0:8000", nil)
	if errhttp != nil {
		fmt.Println(errhttp.Error())
	}
}
