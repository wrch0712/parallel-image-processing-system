package scheduler

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"proj2/png"
	"strings"
	"sync"
	"sync/atomic"
)

// For the BSP model, my design is that we work on 1 image and 1 effect in one super step, each BSPWorker applies effect to a slice of that image
// If all BSPWorkers finishes their work, this super step is finished ,and we go for next super step
// where we apply next effect to that image if needed, or go for the next image

type bspWorkerContext struct {
	config           Config
	dataDirs         []string // big/mixture/small
	taskBuffer       *json.Decoder
	currTask         *png.ImageTask
	currDataDirIndex *int32 // which data directory is being executed now
	currEffectIndex  *int32 // which Effect is being applied now
	isValid          *int32 // weather the current is valid and can be processed, 0 represents for invalid, 1 represents for valid
	stepDoneCount    *int32 // how many goroutines has already finished its task in this super step
	mutex            *sync.Mutex
	cond             *sync.Cond
}

func NewBSPContext(config Config) *bspWorkerContext {
	dataDirs := strings.Split(config.DataDirs, "+")
	var mutex sync.Mutex
	cond := sync.NewCond(&mutex)
	ctx := bspWorkerContext{
		config:           config,
		dataDirs:         dataDirs,
		taskBuffer:       nil,
		currTask:         nil,
		currDataDirIndex: new(int32),
		currEffectIndex:  new(int32),
		isValid:          new(int32),
		stepDoneCount:    new(int32),
		mutex:            &mutex,
		cond:             cond,
	}
	return &ctx
}

func RunBSPWorker(id int, ctx *bspWorkerContext) {
	// The first come goroutine initializes the taskBuffer
	ctx.mutex.Lock()
	if ctx.taskBuffer == nil {
		effectsPathFile := fmt.Sprintf("../data/effects.txt")
		effectsFile, err := os.Open(effectsPathFile)
		if err != nil {
			panic(err)
		}
		defer effectsFile.Close()
		reader := json.NewDecoder(effectsFile)
		ctx.taskBuffer = reader
	}
	ctx.mutex.Unlock()

	for {
		// I choose the main goroutine (id == ctx.config.ThreadCount-1) to be the master
		// It is responsible for generating new tasks
		if id == ctx.config.ThreadCount-1 {
			if ctx.currTask != nil && atomic.LoadInt32(ctx.currEffectIndex) < int32(len(ctx.currTask.Effects))-1 {
				// If there exists other test effect to apply for the same image, apply the effect to the same image in this super step
				atomic.AddInt32(ctx.currEffectIndex, 1)
				atomic.CompareAndSwapInt32(ctx.isValid, 0, 1) // validate the current job
				atomic.StoreInt32(ctx.stepDoneCount, 0)       // reset stepDoneCount
				// wake up workers
				ctx.mutex.Lock()
				ctx.cond.Broadcast()
				ctx.mutex.Unlock()
			} else if ctx.currTask != nil && atomic.LoadInt32(ctx.currDataDirIndex) < int32(len(ctx.dataDirs))-1 {
				// Save the image for previous super step
				err := ctx.currTask.Save()
				if err != nil {
					panic(err)
				}
				// If there are other data directories to deal with for the same image, apply effects to image in that folder
				atomic.AddInt32(ctx.currDataDirIndex, 1)
				oldOutputFilePath := ctx.currTask.OutputFilePath
				ctx.currTask.OutputFilePath = filepath.Join("../data/out", ctx.dataDirs[*ctx.currDataDirIndex]+oldOutputFilePath[strings.Index(oldOutputFilePath, "_"):])
				atomic.CompareAndSwapInt32(ctx.isValid, 0, 1) // validate the current job
				atomic.StoreInt32(ctx.stepDoneCount, 0)       // reset stepDoneCount
				// wake up workers
				ctx.mutex.Lock()
				ctx.cond.Broadcast()
				ctx.mutex.Unlock()
			} else if ctx.taskBuffer.More() {
				// Save the image for previous super step
				if ctx.currTask != nil {
					err := ctx.currTask.Save()
					if err != nil {
						panic(err)
					}
				}
				// Get new task from the task buffer
				var line interface{}
				err := ctx.taskBuffer.Decode(&line)
				if err != nil {
					panic(err)
				}
				inPath := line.(map[string]interface{})["inPath"].(string) // get value form JSON
				outPath := line.(map[string]interface{})["outPath"].(string)
				effects := line.(map[string]interface{})["effects"].([]interface{})
				inputFilePath := filepath.Join("../data/in", ctx.dataDirs[0], inPath)
				outputFilePath := filepath.Join("../data/out", ctx.dataDirs[0]+"_"+outPath)
				ctx.currTask, err = png.Load(inputFilePath) // generate imageTask
				if err != nil {
					panic(err)
				}
				ctx.currTask.Effects = effects
				ctx.currTask.OutputFilePath = outputFilePath
				// validate the current job and reset relative values in ctx
				atomic.StoreInt32(ctx.currDataDirIndex, 0)
				atomic.StoreInt32(ctx.currEffectIndex, 0)
				atomic.StoreInt32(ctx.stepDoneCount, 0)
				atomic.StoreInt32(ctx.isValid, 1)
				// wake up workers
				ctx.mutex.Lock()
				ctx.cond.Broadcast()
				ctx.mutex.Unlock()
			} else {
				// Save the image for previous super step
				if ctx.currTask != nil {
					err := ctx.currTask.Save()
					if err != nil {
						panic(err)
					}
				}
				// No works remaining, end the program
				break
			}
		}

		// if no job is available, workers wait
		ctx.mutex.Lock()
		if ctx.currTask == nil || atomic.LoadInt32(ctx.isValid) == 0 {
			ctx.cond.Wait()
		}
		ctx.mutex.Unlock()

		// Do work: each worker works on a slice of the image
		if ctx.currTask != nil && atomic.LoadInt32(ctx.isValid) == 1 {
			bounds := ctx.currTask.Bounds
			sectionLength := int(math.Ceil(float64(bounds.Max.Y-bounds.Min.Y) / float64(ctx.config.ThreadCount)))
			sectionStart := bounds.Min.Y + sectionLength*id
			if sectionStart > bounds.Max.Y {
				sectionStart = bounds.Max.Y
			}
			sectionEnd := sectionStart + sectionLength + 1
			if sectionEnd > bounds.Max.Y {
				sectionEnd = bounds.Max.Y
			}
			// create mini-imageTask for this worker, each mini-imageTask apply effect to a section of the whole image
			miniImageTask := png.ImageTask{}
			sectionBounds := image.Rect(ctx.currTask.Bounds.Min.X, sectionStart, ctx.currTask.Bounds.Max.X, sectionEnd)
			miniImageTask.In = ctx.currTask.In.SubImage(sectionBounds).(*image.RGBA64)
			miniImageTask.Out = ctx.currTask.Out.SubImage(sectionBounds).(*image.RGBA64)
			miniImageTask.Bounds = sectionBounds

			// Apply effect on the mini-imageTask
			if len(ctx.currTask.Effects) == 0 {
				// If no effects are specified (e.g., []), then out image is the same as the input image.
				miniImageTask.Out = miniImageTask.In
			} else {
				switch ctx.currTask.Effects[*ctx.currEffectIndex] {
				case "S":
					miniImageTask.Sharpen()
					break
				case "E":
					miniImageTask.EdgeDetection()
					break
				case "B":
					miniImageTask.Blur()
					break
				case "G":
					miniImageTask.Grayscale()
					break
				}
				miniImageTask.In = miniImageTask.Out
			}

			ctx.mutex.Lock()
			*ctx.stepDoneCount++
			if *ctx.stepDoneCount == int32(ctx.config.ThreadCount) {
				// If all goroutines has finished their job, this super step has been finished
				*ctx.isValid = 0 // invalidate the current job
				ctx.cond.Broadcast()
			} else {
				// Wait for other goroutines to finish job in this super step
				ctx.cond.Wait()
			}
			ctx.mutex.Unlock()
		}
	}
}
