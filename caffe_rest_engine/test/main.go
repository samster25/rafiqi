package main
// #include <stdlib.h>
// #include "classification.h"
import "C"
import "unsafe"
import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var model C.c_model

func classify(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	buffer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
    var mat C.c_mat
    mat = C.make_mat(model, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)))
    cstr := C.model_classify(model, mat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer C.free(unsafe.Pointer(cstr))
	io.WriteString(w, C.GoString(cstr))
}

func main() {
	cmodel := C.CString(os.Args[1])
	ctrained := C.CString(os.Args[2])
	cmean := C.CString(os.Args[3])
	clabel := C.CString(os.Args[4])

	log.Println("Initializing Caffe classifiers")
	var err error
	model, err = C.model_init(cmodel, ctrained, cmean, clabel)
	if err != nil {
		log.Fatalln("could not initialize classifier:", err)
		return
	}
	defer C.model_destroy(model)
	log.Println("Adding REST endpoint /api/classify")
	http.HandleFunc("/api/classify", classify)
	log.Println("Starting server listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
