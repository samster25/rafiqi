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
	InGPU      bool
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
	Debugf("Currently %d entries in LRU", g.LRU.Len())
	g.LRULock.Lock()
	evicted := g.LRU.Back()

	if evicted == nil {
		fmt.Println("Nothing in GPU!")
		time.Sleep(10 * time.Second)
		panic("Exceeded mem usage, but no models loaded!")
	}

	model := (evicted.Value).(Model)
	/*
		// Make sure we aren't evicting ourselves
		if model.Name == m.Name {
			evicted = evicted.Prev()
			if evicted == nil {
				panic("")
			}
		}
	*/
	g.LRU.Remove(evicted)
	g.LRULock.Unlock()

	Debugf("%v is the lucky victim", model.Name)

	entry, ok := loadedModels.Models[model.Name]
	if !ok {
		panic("Tried to evict model not in loaded models: " + model.Name)
	} else if !entry.InGPU {
		panic("Tried to evict model not in GPU: " + model.Name)
	}

	entry.InGPU = false
	Debugf("Destroy beginning!")
	start := time.Now()
	//C.model_destroy(entry.Classifier)
	g.MoveToCPU(&model, entry)
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
	cclass, err := C.model_init(cmodel, cweights, cmean, clabel,
		C.size_t(NUM_CONTEXTS), C.size_t(MAX_BATCH_AMT))
	LogTimef("%v model_init", start, m.Name)

	if err != nil {
		handleError("init failed: ", err)
	}

	C.move_to_cpu(cclass)
	C.move_to_gpu(cclass)
	g.LRULock.Lock()
	Debugf("Adding to LRU: %v", m.Name)
	g.LRU.PushBack(*m)
	g.LRULock.Unlock()

	return &ModelEntry{
		Classifier: cclass,
		InGPU:      true,
	}
}

func (g *GPUMem) UpdateLRU(m *Model) {
	Debugf("In update LRU %v", m.Name)
	g.LRULock.Lock()
	defer g.LRULock.Unlock()
	for e := g.LRU.Front(); e != nil; e = e.Next() {
		model := (e.Value).(Model)
		if model.Name == m.Name {
			g.LRU.MoveToFront(e)
			return
		}
	}

	panic("Tried to update model, but not in LRU: " + m.Name)
}

func (g *GPUMem) MoveToCPU(m *Model, entry *ModelEntry) {
	start := time.Now()
	Debugf("%v move to cpu beginning", m.Name)
	C.move_to_cpu(entry.Classifier)
	LogTimef("%v move to cpu", start, m.Name)
}

func (g *GPUMem) MoveToGPU(m *Model, entry *ModelEntry) {
	Debugf("About to move %v onto the GPU", m.Name)
	g.InitLock.Lock()
	for !g.CanLoad(m) {
		g.EvictLRU()
	}
	g.InitLock.Unlock()

	start := time.Now()
	entry.InGPU = true
	C.move_to_gpu(entry.Classifier)
	g.LRULock.Lock()
	g.LRU.PushFront(*m)
	g.LRULock.Unlock()

	LogTimef("%v move to gpu", start, m.Name)
}

func (g *GPUMem) LoadModel(m Model) *ModelEntry {
	Debugf("LoadModel begin for %v", m.Name)
	Debugf("Currently %d entries in LRU", g.LRU.Len())
	var entry *ModelEntry
	loadedModels.RLock()
	entry, ok := loadedModels.Models[m.Name]
	loadedModels.RUnlock()

	if !ok {
		loadedModels.Lock()
		// Ensure no one added this model between the RUnlock and here
		_, ok = loadedModels.Models[m.Name]
		if !ok {
			entry = g.InitModel(&m)
			loadedModels.Models[m.Name] = entry
		}
		loadedModels.Unlock()
	} else if !entry.InGPU {
		g.MoveToGPU(&m, entry)
	}

	g.UpdateLRU(&m)
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
