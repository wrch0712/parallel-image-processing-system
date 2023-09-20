import matplotlib
import pandas as pd
import matplotlib.pyplot as plt
import warnings
warnings.filterwarnings("ignore")
matplotlib.use('agg')
'''
makeGraphs.py makes speedup graph based on data in speedup.csv 
and save the graph as speedupGraph.png
'''
# Make speedup graph for pipeline model
# Read CSV file
df_pipline = pd.read_csv('speedup_pipeline.csv')

# Extract test cases and threads
test_cases = df_pipline['Test Case']
threads = df_pipline.columns[1:]

# Loop through test cases and generate speedup plot
for test_case in test_cases:
    # Extract speedup values for the corresponding test case
    speedup_values = df_pipline[df_pipline['Test Case'] == test_case].values[0][1:]
    # Make plot
    plt.plot(threads, speedup_values, label=test_case)

# Add plot title and labels
plt.title('Speedup-Pipeline')
plt.xlabel('Number of Threads')
plt.ylabel('Speedup')
# Add legend
plt.legend()

# Save plot as image file
plt.savefig('speedup-pipeline.png')

plt.clf()
# Make speedup graph for bsp model
# Read CSV file
df_bsp = pd.read_csv('speedup_bsp.csv')

# Extract test cases and threads
test_cases = df_bsp['Test Case']
threads = df_bsp.columns[1:]

# Loop through test cases and generate speedup plot
for test_case in test_cases:
    # Extract speedup values for the corresponding test case
    speedup_values = df_bsp[df_bsp['Test Case'] == test_case].values[0][1:]
    # Make plot
    plt.plot(threads, speedup_values, label=test_case)

# Add plot title and labels
plt.title('Speedup-BSP')
plt.xlabel('Number of Threads')
plt.ylabel('Speedup')
# Add legend
plt.legend()

# Save plot as image file
plt.savefig('speedup-bsp.png')
