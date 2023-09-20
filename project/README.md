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
|  ``in`` directory | This directory contains three subdirectories called: ``big``, ``mixture``, and ``small``. The actual images in each of these subdirectories are all the same, with the exception of their *image sizes*. The ``big`` directory has the best resolution of the images, ``small`` has a reduced resolution of the images, and the ``mixture`` directory has a mixture of both big and small sizes for different images. You must use a relative path to your ``proj2`` directory to open this file. For example, if you want to open the ``IMG_2029.png`` from the ``big`` directory from inside the ``editor.go`` file then you should open as ``../data/in/big/IMG_2029.png``. |
| ``out`` directory | This is where will place the ``outPath`` images when running the program. |
