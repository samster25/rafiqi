// gpu_cache.cpp
#include "gpu_cache.h"

GPU_Cache::GPU_Cache(int size) {
    max_size = size;
}

int GPU_Cache::dim_to_key(int m, int n) {
    return m + (n << 32);
}

bool GPU_Cache::get(int m, int n, void **cpu_ptr, void **gpu_ptr) {
    int key = dim_to_key(m, n);
    if (cache_map.count(key) > 0) {
        cache_block *found = cache_map[key];
        *gpu_ptr = found->gpu_ptr;
        *cpu_ptr = found->cpu_ptr;
        return true;
    } else {
        return false;
    }
}

bool GPU_Cache::put(int m, int n, void **cpu_ptr, void **gpu_ptr) {
    return true;
}

bool GPU_Cache::put_and_malloc(int m, int n, void **cpu_ptr, void **gpu_ptr) {
    return true;
}
