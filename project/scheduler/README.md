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
