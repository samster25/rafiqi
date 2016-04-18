data := /work/matrix_models
manifesto := /work/matrix_models/manifesto.txt.example
manifesto_used := /work/matrix_models/manifesto.txt
stats := ./data/$(shell date +%s)_data.csv
start := 129
end := 16384 
step := 2048


generate:
	python gen_matrix.py $(data) $(manifesto) $(start) $(end) $(step)
	cp $(manifesto) $(manifesto_used)
build:
	mkdir -p bin
	nvcc -lcublas -O3 gpu_benchmark.cpp gpu_cache.cpp -std=c++11 -o bin/gpu_benchmark
randomPattern:
	python generate_access_patterns.py $(manifesto) $(manifesto_used) random
standardPattern:
	python generate_access_patterns.py $(manifesto) $(manifesto_used) standard
benchmark:
	mkdir -p data
	./bin/gpu_benchmark $(data) $(manifesto_used) $(stats)
