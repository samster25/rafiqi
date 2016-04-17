data := /work/matrix_models
manifesto := /work/matrix_models/manifesto.txt
stats := ./data/$(shell date +%s)_data.csv

start := 128
end := 16384 
step := 2048


generate:
	/usr/bin/python gen_matrix.py $(data) $(manifesto) $(start) $(end) $(step)
build:
	mkdir -p bin
	nvcc -lcublas -O3 gpu_benchmark.cpp gpu_cache.cpp -std=c++11 -o bin/gpu_benchmark
benchmark:
	mkdir -p data
	./bin/gpu_benchmark $(data) $(manifesto) $(stats)
