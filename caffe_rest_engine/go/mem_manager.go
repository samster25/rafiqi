package main

// #include <stdlib.h>
// #include <classification.h>
import "C"
import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type GPUMem struct {
	InitLock sync.Mutex
	LRU      *list.List
	LRULock  sync.Mutex
}

type LoadedModelsMap struct {
	sync.RWMutex
	Models map[string]*ModelEntry
}

type ModelEntry struct {
	sync.Mutex
	Classifier C.c_model
}

var loadedModels LoadedModelsMap
var MemoryManager *GPUMem

func (g *GPUMem) GetCurrentMemUsage() uint64 {
	return uint64(C.get_total_gpu_memory()) - uint64(C.get_free_gpu_memory())
}

func (g *GPUMem) CanLoad(m *Model) bool {
	return g.GetCurrentMemUsage()+m.estimatedGPUMemSize() < maxGPUMemUsage
}

func (g *GPUMem) EvictLRU() {
	Debugf("Eviction beginning!")
	g.LRULock.Lock()
	evicted := g.LRU.Back()

	if evicted == nil {
		panic("Exceeded mem usage, but no models loaded!")
	}

	model := (evicted.Value).(*Model)
	g.LRU.Remove(evicted)
	g.LRULock.Unlock()

	Debugf("%v is the lucky victim", model.Name)

	entry, ok := loadedModels.Models[model.Name]
	if !ok {
		panic("Tried to evict model not in loaded models: " + model.Name)
	}
	delete(loadedModels.Models, model.Name)

	Debugf("Destroy beginning!")
	start := time.Now()
	C.model_destroy(entry.Classifier)
	LogTimef("%v model destroy", start, model.Name)
}

func (g *GPUMem) InitModel(m *Model) *ModelEntry {
	Debugf("Initializing model: %v", m.Name)
	Debugf("Current mem usage: %d mebibytes", g.GetCurrentMemUsage()/(1024*1024))
	g.InitLock.Lock()
	for !g.CanLoad(m) {
		g.EvictLRU()
	}
	g.InitLock.Unlock()
	cmean := C.CString(m.MeanPath)
	clabel := C.CString(m.LabelsPath)
	cweights := C.CString(m.WeightsPath)
	cmodel := C.CString(m.ModelPath)
	start := time.Now()
	cclass, err := C.model_init(cmodel, cweights, cmean, clabel)
	fmt.Println("here", m.Name)
	LogTimef("%v model_init", start, m.Name)

	if err != nil {
		handleError("init failed: ", err)
	}

	g.LRULock.Lock()
	g.LRU.PushBack(m)
	g.LRULock.Unlock()

	return &ModelEntry{Classifier: cclass}
}

func (g *GPUMem) UpdateLRU(m *Model) {
	g.LRULock.Lock()
	defer g.LRULock.Unlock()
	for e := g.LRU.Front(); e != nil; e = e.Next() {
		model := (e.Value).(*Model)
		if model.Name == m.Name {
			g.LRU.MoveToFront(e)
			break
		}
	}
}

func (g *GPUMem) LoadModel(m *Model) *ModelEntry {
	Debugf("LoadModel begin")
	var entry *ModelEntry
	loadedModels.RLock()
	entry, ok := loadedModels.Models[m.Name]
	loadedModels.RUnlock()

	if !ok {

		loadedModels.Lock()
		// Ensure no one added this model between the RUnlock and here
		_, ok = loadedModels.Models[m.Name]
		if !ok {
			entry = g.InitModel(m)
			loadedModels.Models[m.Name] = entry
		}
		loadedModels.Unlock()
	}

	g.UpdateLRU(m)
	return entry

}

func init() {
	loadedModels = LoadedModelsMap{
		Models: make(map[string]*ModelEntry),
	}

	MemoryManager = &GPUMem{
		LRU: list.New(),
	}

}
