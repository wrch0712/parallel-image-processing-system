#!/bin/bash

# runs.sh will execute all required runs,
# and make a new file called speedup.csv which stores the speedup value for each test case and each thread number
# and call makeGraphs.py to make graph based on the result in speedup.csv

# Define the test cases and number of threads to iterate over
test_cases=("small" "mixture" "big")
threads=(2 4 6 8 12)

# Create the CSV files with headers
echo "Test Case,2,4,6,8,12" > speedup_pipeline.csv
echo "Test Case,2,4,6,8,12" > speedup_bsp.csv

# Loop through the test cases
for test_case in "${test_cases[@]}"; do
  # get sequential elapsed time
  sum_s=0
  # Run sequential version 5 times and capture the elapsed times
  for ((i=1; i<=5; i++)); do
    output=$(go run ../editor/editor.go "$test_case")
    elapsed_time=$(echo "$output" | awk '{print $1}')
    sum_s=$(echo "$sum_s + $elapsed_time" | bc)
  done
  # Calculate the average sequential elapsed time
  avg_s=$(echo "scale=3; $sum_s / 5" | bc)

  # Experiment with the pipeline model
  pipeline_speedup_arr=() # an array to store the speedup values for this test case
  # get parallel elapsed time for each thread number
  for thread in "${threads[@]}"; do
    sum_p=0
    # Run parallel version 5 times with certain thread number and capture the elapsed times
    for ((i=1; i<=5; i++)); do
      output=$(go run ../editor/editor.go "$test_case" pipeline "$thread")
      elapsed_time=$(echo "$output" | awk '{print $1}')
      sum_p=$(echo "$sum_p + $elapsed_time" | bc)
    done
    # Calculate the average parallel elapsed time
    avg_p=$(echo "scale=3; $sum_p / 5" | bc)

    # Calculate the speedup
    speedup=$(echo "scale=3; $avg_s / $avg_p" | bc)
    # Append the speedup value to the array
    pipeline_speedup_arr+=("$speedup")
  done

  # Output the pipeline speedup values to the CSV file
  echo "$test_case,${pipeline_speedup_arr[0]},${pipeline_speedup_arr[1]},${pipeline_speedup_arr[2]},${pipeline_speedup_arr[3]},${pipeline_speedup_arr[4]}" >> speedup_pipeline.csv

  # Experiment with the bsp model
  bsp_speedup_arr=() # an array to store the speedup values for this test case
  # get parallel elapsed time for each thread number
  for thread in "${threads[@]}"; do
    sum_p=0
    # Run parallel version 5 times with certain thread number and capture the elapsed times
    for ((i=1; i<=5; i++)); do
      output=$(go run ../editor/editor.go "$test_case" bsp "$thread")
      elapsed_time=$(echo "$output" | awk '{print $1}')
      sum_p=$(echo "$sum_p + $elapsed_time" | bc)
    done
    # Calculate the average parallel elapsed time
    avg_p=$(echo "scale=3; $sum_p / 5" | bc)

    # Calculate the speedup
    speedup=$(echo "scale=3; $avg_s / $avg_p" | bc)
    # Append the speedup value to the array
    bsp_speedup_arr+=("$speedup")
  done

  # Output the pipeline speedup values to the CSV file
  echo "$test_case,${bsp_speedup_arr[0]},${bsp_speedup_arr[1]},${bsp_speedup_arr[2]},${bsp_speedup_arr[3]},${bsp_speedup_arr[4]}" >> speedup_bsp.csv
done

# call makeGraphs.py to generate the speedup graph based on result in csv files
python3 makeGraphs.py
