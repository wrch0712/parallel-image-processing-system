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
)

// Pipeline model implements the fan-in/fan-out scheme.
// ImageTaskGenerator produces ImageTasks and dumps them into an imageTask channel,
// workers try to grab ImageTasks from the channel. A worker is solely responsible for performing all effects for one image
// and it will spawn mini-workers which apply effects on a slice of the image.
// A resultsAggregator aggregates the ImageResult from workers’own ImageResults channels and returns a single channel of ImageResult structs, the ImageResult will be saved into the output path

func RunPipeline(config Config) {
	// Initialize the taskBuffer (json decoder)
	effectsPathFile := fmt.Sprintf("../data/effects.txt")
	effectsFile, err := os.Open(effectsPathFile)
	if err != nil {
		panic(err)
	}
	defer effectsFile.Close()
	reader := json.NewDecoder(effectsFile)

	done := make(chan interface{})
	defer close(done)

	n := config.ThreadCount
	imageTasksStream := imageTaskGenerator(reader, done, config)
	workerImageResultsStreams := make([]<-chan *png.ImageTask, n)
	// spawn n workers
	for i := 0; i < n; i++ {
		workerImageResultsStreams[i] = worker(done, imageTasksStream, n)
	}
	imageResultsStream := resultsAggregator(done, workerImageResultsStreams)
	for imageResult := range imageResultsStream {
		err = imageResult.Save()
		if err != nil {
			panic(err)
		}
	}
}

// The imageTaskGenerator reads tasks from a JSON decoder and generates imageTasks based on the configuration.
// It returns a imageTasksStream channel and each imageTask generated is put imageTasksStream channel.
func imageTaskGenerator(reader *json.Decoder, done <-chan interface{}, config Config) <-chan *png.ImageTask {
	imageTasksStream := make(chan *png.ImageTask)
	// Read tasks from data/effects.txt
	dataDirs := strings.Split(config.DataDirs, "+") // get data directories from config

	var line interface{}
	go func() {
		defer close(imageTasksStream)
		for reader.More() {
			select {
			case <-done:
				return
			default:
				err := reader.Decode(&line)
				if err != nil {
					panic(err)
				}
				// get value form JSON
				inPath := line.(map[string]interface{})["inPath"].(string)
				outPath := line.(map[string]interface{})["outPath"].(string)
				effects := line.(map[string]interface{})["effects"].([]interface{})

				// Loop over every mode in the configuration
				for _, dataDir := range dataDirs {
					inputFilePath := filepath.Join("../data/in", dataDir, inPath)
					outputFilePath := filepath.Join("../data/out", dataDir+"_"+outPath)
					pngImg, err := png.Load(inputFilePath) // generate imageTask
					if err != nil {
						panic(err)
					}
					pngImg.Effects = effects
					pngImg.OutputFilePath = outputFilePath
					// put image task to the output imageTaskStream
					imageTasksStream <- pngImg
				}
			}
		}
	}()
	return imageTasksStream
}

// Each worker goroutine grabs imageTasks from the input imageTasksStream channel.
// Then worker slices the image into n piece (generate n miniImageTasks) and spawn n mini-workers to deal with miniImageTasks
// After all mini-workers finishes, worker aggregate their works, reassemble into a single imageResult and put in its output workerImageResultsStream channel.
func worker(done <-chan interface{}, imageTasksStream <-chan *png.ImageTask, n int) <-chan *png.ImageTask {
	workerImageResultsStream := make(chan *png.ImageTask)
	go func() {
		defer close(workerImageResultsStream)
		for imageTask := range imageTasksStream {
			select {
			case <-done:
				return
			default:
				// case imageTask := <-imageTasksStream:
				if len(imageTask.Effects) == 0 {
					// If no effects are specified (e.g., []), then out image is the same as the input image.
					imageTask.Out = imageTask.In
					workerImageResultsStream <- imageTask
				} else {
					// Separate the imageTask into n mini-imageTask and store a mini-imageTasks in a slice
					// Each mini-imageTask correspond to an image section that a mini-worker need to deal with
					miniTasks := make([]*png.ImageTask, n)
					// calculate the length of the image section that each mini-worker need to deal with
					bounds := imageTask.Bounds
					sectionLength := int(math.Ceil(float64(bounds.Max.Y-bounds.Min.Y) / float64(n)))
					for i := 0; i < n; i++ {
						sectionStart := bounds.Min.Y + sectionLength*i
						if sectionStart > bounds.Max.Y {
							sectionStart = bounds.Max.Y
						}
						sectionEnd := sectionStart + sectionLength + 1
						if sectionEnd > bounds.Max.Y {
							sectionEnd = bounds.Max.Y
						}
						// create mini-imageTask for mini-worker based on the image section
						miniImageTask := png.ImageTask{}
						sectionBounds := image.Rect(imageTask.Bounds.Min.X, sectionStart, imageTask.Bounds.Max.X, sectionEnd)
						miniImageTask.In = imageTask.In.SubImage(sectionBounds).(*image.RGBA64)
						miniImageTask.Out = imageTask.Out.SubImage(sectionBounds).(*image.RGBA64)
						miniImageTask.Bounds = sectionBounds

						// put the mini-imageTask in the mini-task slice
						miniTasks[i] = &miniImageTask
					}

					// Spawns mini-workers to apply effects to the image section
					for _, effect := range imageTask.Effects {
						// make a waitGroup using channel
						// because we need to wait until all mini-worker have applied one effect to apply another effect
						wgChan := make(chan int, n)
						// spread out n mini-worker to apply the effect for its assigned section.
						for i := 0; i < n; i++ {
							miniTask := miniTasks[i]
							go func() { // a mini-worker
								// Apply the specified effect
								switch effect {
								case "S":
									miniTask.Sharpen()
									break
								case "E":
									miniTask.EdgeDetection()
									break
								case "B":
									miniTask.Blur()
									break
								case "G":
									miniTask.Grayscale()
									break
								}
								// Swap the pointers to make the old out buffer the new in buffer
								// when applying one effect after another effect
								miniTask.In = miniTask.Out
								wgChan <- 1 //wg.done()
							}()
						}
						for i := 0; i < n; i++ {
							<-wgChan // wg.wait()
						}
					}

					// Put the final image after applying its effects to output imageResultStream
					workerImageResultsStream <- imageTask
				}
			}
		}
	}()
	return workerImageResultsStream
}

// The resultsAggregator aggregates the ImageResult from n workerImageResultsStream channels
// and returns a single channel of ImageResult structs.
func resultsAggregator(done <-chan interface{}, workerImageResultsStreams []<-chan *png.ImageTask) <-chan *png.ImageTask {
	imageResultsStream := make(chan *png.ImageTask)
	wgChan := make(chan int, len(workerImageResultsStreams))

	helper := func(workerImageResultsStream <-chan *png.ImageTask) {
		defer func() {
			wgChan <- 1 //wg.done()
		}()
		for imageResult := range workerImageResultsStream {
			select {
			case <-done:
				return
			case imageResultsStream <- imageResult:
			}
		}
	}

	for _, workerImageResultsStream := range workerImageResultsStreams {
		go helper(workerImageResultsStream)
	}

	// Wait for all the reads to complete
	go func() {
		for i := 0; i < len(workerImageResultsStreams); i++ {
			<-wgChan // wg.wait()
		}
		close(imageResultsStream)
	}()
	return imageResultsStream
}
