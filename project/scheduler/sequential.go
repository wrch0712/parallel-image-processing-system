package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"proj2/png"
	"strings"
)

func RunSequential(config Config) {
	effectsPathFile := fmt.Sprintf("../data/effects.txt")
	effectsFile, err := os.Open(effectsPathFile)
	if err != nil {
		panic(err)
	}
	defer effectsFile.Close()

	reader := json.NewDecoder(effectsFile)

	// get data directories from config
	dataDirs := strings.Split(config.DataDirs, "+")

	var line interface{}
	// read each line
	for reader.More() {
		err = reader.Decode(&line)
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
			pngImg, err := png.Load(inputFilePath)
			if err != nil {
				panic(err)
			}
			pngImg.OutputFilePath = outputFilePath
			// Apply the specified effects to the input image
			if len(effects) == 0 {
				// If no effects are specified (e.g., [])
				// then out image is the same as the input image.
				pngImg.Out = pngImg.In
			} else {
				for _, effect := range effects {
					// Apply the specified effect
					switch effect {
					case "S":
						pngImg.Sharpen()
						break
					case "E":
						pngImg.EdgeDetection()
						break
					case "B":
						pngImg.Blur()
						break
					case "G":
						pngImg.Grayscale()
						break
					}

					// Swap the pointers to make the old out buffer the new in buffer
					// when applying one effect after another effect
					pngImg.In = pngImg.Out
				}
			}
			//Saves the image
			err = pngImg.Save()
			if err != nil {
				panic(err)
			}
		}
	}
}
