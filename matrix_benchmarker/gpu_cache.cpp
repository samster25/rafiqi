// gpu_cache.cpp
#include "gpu_cache.h"
#include <cuda.h>
#include <cuda_runtime.h>

GPU_Cache::GPU_Cache(int size) {
    max_size = size;
}

int GPU_Cache::dim_to_key(int m, int n) {
    return m + (n << 32);
}

bool GPU_Cache::get(int m, int n, void **cpu_ptr, void **gpu_ptr) {
    int key = dim_to_key(m, n);
    if (cache_map.count(key) > 0) {
        cache_block found = cache_map[key];
        *gpu_ptr = found.gpu_ptr;
        *cpu_ptr = found.cpu_ptr;
        return true;
    } else {
        return false;
    }
}

bool GPU_Cache::put_and_malloc(int m, int n, void **cpu_ptr, void **gpu_ptr) {
    
    if ((cudaMalloc((void **) gpu_ptr, m*n*sizeof(float)) != cudaSuccess)) {
        printf("error cuda mallocing\n");
        exit(1);
    }
    cudaMemcpy(gpu_ptr, cpu_ptr, m*n*sizeof(float), cudaMemcpyHostToDevice);
    return put(m, n, cpu_ptr, gpu_ptr);
}

bool GPU_Cache::put(int m, int n, void **cpu_ptr, void **gpu_ptr) {
    cache_block *curr = (* cache_block)malloc(sizeof(cache_block));
    curr->gpu_ptr = *gpu_ptr;
    curr->cpu_ptr = *cpu_ptr;
    curr->m = m;
    curr->n = n;
    cache_map[dim_to_key(m, n)] = *curr;
    return true;
}

int main (int argc, char **argv) {
    return 0;
}
