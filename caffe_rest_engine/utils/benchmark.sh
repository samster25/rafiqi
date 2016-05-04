mkdir -p benchmarks
mkdir -p plots

echo "Running benchmarks for 1 model/1 concurrent request"
ab -n 1024 -c 1 -T 'application/x-www-form-urlencoded' -p cat.jpg -e benchmarks/data_1_1_caffenet.csv -g plots/data_1_1_caffenet.tsv "$2=caffenet"
ab -n 1024 -c 1 -T 'application/x-www-form-urlencoded' -p cat.jpg -e benchmarks/data_1_1_googlenet.csv -g plots/data_1_1_googlenet.tsv "$2=googlenet"
ab -n 1024 -c 1 -T 'application/x-www-form-urlencoded' -p cat.jpg -e benchmarks/data_1_1_alexnet.csv -g plots/data_1_1_alexnet.tsv "$2=alexnet"

echo "Running benchmarks for 1 model/n concurrent reqs"
ab -n 1024 -c $1 -T 'application/x-www-form-urlencoded' -p cat.jpg -e benchmarks/data_1_$1\_caffenet.csv -g plots/data_1_$1\_caffenet.tsv "$2=caffenet"
ab -n 1024 -c $1 -T 'application/x-www-form-urlencoded' -p cat.jpg -e benchmarks/data_1_$1\_googlenet.csv  -g plots/data_1_$1\_googlenet.tsv "$2=googlenet"
ab -n 1024 -c $1 -T 'application/x-www-form-urlencoded' -p cat.jpg -e benchmarks/data_1_$1\_alexnet.csv -g plots/data_1_$1\_alexnet.tsv "$2=alexnet"

echo "Running batching benchmarks"
ab -n 1024 -c 32 -T 'application/x-www-form-urlencoded' -p cat.jpg -e benchmarks/data_batch_caffenet.csv  -g plots/data_batch_caffenet.tsv "$2=caffenet"
ab -n 1024 -c 32 -T 'application/x-www-form-urlencoded' -p cat.jpg -e benchmarks/data_batch_googlenet.csv  -g plots/data_batch_googlenet.tsv "$2=googlenet"
ab -n 1024 -c 32 -T 'application/x-www-form-urlencoded' -p cat.jpg -e benchmarks/data_batch_alexnet.csv -g plots/data_1_batch_alexnet.tsv "$2=alexnet"
echo "Multiple models at a time"

mkdir -p plot_images
for filename in /plots/*.tsv; do
    printf "set terminal jpeg size 500,500 \n
    # This sets the aspect ratio of the graph \n
    set size 1, 1 \n
    # The file we'll write to \n
    set output "plot_images/$(basename "$filename")\.jpg"\n
    # The graph title \n
    set title "Benchmark testing"\n
    # Where to place the legend/key \n
    set key left top \n
    # Draw gridlines oriented on the y axis \n
    set grid y \n
    # Specify that the x-series data is time data\n
    set xdata time\n
    # Specify the *input* format of the time data\n
    set timefmt "%s"\n
    # Specify the *output* format for the x-axis tick labels\n
    set format x "%S"\n
    # Label the x-axis\n
    set xlabel 'seconds'\n
    # Label the y-axis\n
    set ylabel "response time (ms)"\n
    # Tell gnuplot to use tabs as the delimiter instead of spaces (default)\n
    set datafile separator '\t'\n
    # Plot the data\n
    plot "$filename" every ::2 using 2:5 title 'response time' with points\n
    exit\n" > plot.plt
    gnuplot plot.plt

