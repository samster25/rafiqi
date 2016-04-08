data := /scratch/sammy/matrix_models
manifesto := /scratch/sammy/matrix_models/manifesto.txt
stats := ./data.csv

start := 9
end := 14
step := 1

generate:
	python gen_matrix.py $(data) $(manifesto) $(start) $(end) $(step)
build:
	nvcc -lcublas -O3 gpu_benchmark.cpp -o bin/gpu_benchmark
benchmark:
	./bin/gpu_benchmark $(data) $(manifesto) $(stats)
