# Parallel Image Processing System Project

A parallel image processing system that applies image effects on series of images using 2D image convolution.

### The `png` Directory
In png package, ImageTask struct is defined in png.go, and effects.go provides various functions to apply sharpen, edge detection, blur, and grayscale filtering effects to the image based on 2D image convolution

### The `scheduler` Directory
Scheduler package provides sequential and parallel (pipeline, BSP) ways to implement image processing jobs on series of images. 

The program will read images from a series of JSON strings, apply the effects associated with an image, and save the images to their specified output file paths. 

Specifically, the pipeline model implements the fan-in/fan-out scheme. ImageTaskGenerator produces ImageTasks and dumps them into an imageTask channel, workers try to grab ImageTasks from the channel. A worker is solely responsible for performing all effects for one image and it will spawn mini-workers which apply effects on a slice of the image. A resultsAggregator aggregates the ImageResult from workers’ own ImageResults channels and returns a single channel of ImageResult structs, the ImageResult will be saved into the output path. 

For the BSP model, 1 image and 1 effect is implemented in one super step, each worker applies the effect to a slice of that image. If all workers finish their work, this super step is finished, then they go for the next super step. scheduler.go provides configuration, and acts as the entrance of the program.

### The `data` Directory

Here is the structure of the `data` directory:

| Directory/Files | Description  |
|-----------------|--------------|
| ``effects.txt`` |  This is the file that contains the string of JSONS that were described above. This will be the only file used for this program (and also for testing purposes).|
|  ``expected`` directory | This directory contains the expected filtered out image for each JSON string provided in the ``effects.txt``. We will only test your program against the images provided in this directory. Your  produced images do not need to look 100% like the provided output. If there are some slight differences based on rounding-error then that's fine for full credit. |
|  ``in`` directory | This directory contains three subdirectories called: ``big``, ``mixture``, and ``small``. The actual images in each of these subdirectories are all the same, with the exception of their *image sizes*. The ``big`` directory has the best resolution of the images, ``small`` has a reduced resolution of the images, and the ``mixture`` directory has a mixture of both big and small sizes for different images. You must use a relative path to your ``proj2`` directory to open this file. For example, if you want to open the ``IMG_2029.png`` from the ``big`` directory from inside the ``editor.go`` file then you should open as ``../data/in/big/IMG_2029.png``. |
| ``out`` directory | This is where will place the ``outPath`` images when running the program. |


### How to run testing script
To run a single job, go to the scheduler directory and run scheduler.go, which is the entrance of the problem.

usage statement:

    Usage: editor data_dir [mode] [number_of_threads]
    data_dir = The data directories to use to load the images.
    mode     = (bsp) run the BSP mode, (pipeline) run the pipeline mode
    number_of_threads = Runs the parallel version of the program with the specified number of threads (i.e., goroutines).

For example, running the program as follows:

    $: go run editor.go big pipeline 4

To create speedup graphs, go to benchmark directory, and run the runs.sh file. runs.sh does experiment, stores results in speedup_pipeline.csv and speedup_bsp.csv, and calls makeGraphs.py to make speedup graphs based on results in csv files. After running the runs.sh file, you will find speedup_pipeline.csv, speedup_bsp.csv, speedup-pipeline.png, and speedup-bsp.png in the benchmark directory.

### Performance analysis
Please refer to project_report.pdf
