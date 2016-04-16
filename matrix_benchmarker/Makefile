data := /work/matrix_models
manifesto := /work/matrix_models/manifesto.txt
stats := ./data.csv

start := 9
end := 15
step := 1

generate:
	python gen_matrix.py $(data) $(manifesto) $(start) $(end) $(step)
build:
	mkdir -p bin
	nvcc -lcublas -O3 gpu_benchmark.cpp -o bin/gpu_benchmark
benchmark:
	./bin/gpu_benchmark $(data) $(manifesto) $(stats)
