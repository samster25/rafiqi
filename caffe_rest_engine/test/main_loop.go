package main
// #include <stdlib.h>
// #include "classification.h"
import "C"
import "unsafe"
import (
	"io/ioutil"
	"log"
	"os"
    "fmt"
    "runtime"
)


func f(model C.c_model, img_file string) {
    fmt.Println("   IM A FUCKER\n\n\n")
    buffer, _ := ioutil.ReadFile(img_file)
    var mat C.c_mat
    for i := 0; i < 200; i++ {
        mat = C.make_mat(model, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)))
        cstr := C.model_classify(model, mat)
        fmt.Println(*cstr)
	//    defer C.free(unsafe.Pointer(cstr))
    }
}


func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    cmodel := C.CString(os.Args[1])
	ctrained := C.CString(os.Args[2])
	cmean := C.CString(os.Args[3])
	clabel := C.CString(os.Args[4])
    img_file := os.Args[5]
    var model C.c_model
	log.Println("Initializing Caffe classifiers", cmodel)
	var err error
	model, err = C.model_init(cmodel, ctrained, cmean, clabel)
	if err != nil {
		log.Fatalln("could not initialize classifier:", err)
		return
	}

    go f(model, img_file)
    for {} 
}
