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
	"net/http"
)

var (
	nworkers = flag.Int("n", 4, "Enter the number of workers wanted.")
)

func preload() {
	var modelGob []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS_BUCKET)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			modelGob = make([]byte, len(v))
			copy(modelGob, v)
			buf := bytes.NewBuffer(modelGob)
			dec := gob.NewDecoder(buf)
			var model Model
			dec.Decode(&model)
			InitializeModel(&model)
		}
		return nil
	})

	if err != nil {
		panic("error in transaction! " + err.Error())
	}
}

func main() {

	flag.Parse()
	fmt.Println("Preloading and pre-init'ing models")
	preload()
	fmt.Println("Finished prefetching models into CPU Ram")
	fmt.Println("Starting the dispatcher!")
	fmt.Println("nworker", *nworkers)
	dis := NewDispatcher("placeholder", *nworkers)
	dis.StartDispatcher()
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
