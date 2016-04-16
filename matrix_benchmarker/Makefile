data := /scratch/sammy/matrix_models
manifesto := /scratch/sammy/matrix_models/manifesto.txt
stats := ./data/$(shell date +%s)_data.csv

start := 128
end := 16384 
step := 2048


generate:
	python gen_matrix.py $(data) $(manifesto) $(start) $(end) $(step)
build:
	mkdir -p bin
	nvcc -lcublas -O3 gpu_benchmark.cpp -o bin/gpu_benchmark
benchmark:
	mkdir -p data
	./bin/gpu_benchmark $(data) $(manifesto) $(stats)
