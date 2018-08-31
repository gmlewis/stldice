// Package stl provides some useful STL-handling utilities.
package stl

import (
	"log"
	"sync"

	gl "github.com/fogleman/fauxgl"
)

// Merge merges the STL files using a worker pool of numWorkers.
func Merge(files []string, numWorkers int) (mesh *gl.Mesh, err error) {
	var wg sync.WaitGroup
	ch := make(chan struct{}, numWorkers)

	var mu sync.Mutex // protects mesh
	addFunc := func(meshToAdd *gl.Mesh) {
		mu.Lock()
		if mesh == nil {
			mesh = meshToAdd
		} else {
			mesh.Add(meshToAdd)
		}
		mu.Unlock()
	}

	var mergeErr error
	for _, f := range files {
		ch <- struct{}{} // Don't exceed numWorkers goroutines at a time.
		wg.Add(1)
		go func(filename string) {
			if err := merge(filename, addFunc); err != nil {
				mergeErr = err
			}
			<-ch // release a worker
			wg.Done()
		}(f)
	}
	wg.Wait()

	if mergeErr != nil {
		return nil, mergeErr
	}

	return mesh, nil
}

// merge merges STL files using the provided addFunc.
func merge(filename string, addFunc func(mesh *gl.Mesh)) error {
	mesh, err := gl.LoadSTL(filename)
	if err != nil {
		return err
	}
	log.Printf("%v bounding box: %v", filename, mesh.BoundingBox())
	addFunc(mesh)
	return nil
}
