// gpu_cache.h
// #ifndef gpu_cache_h
// #define gpu_cache_h

#include <unordered_map> 
#include <iostream>
#include <string>

struct cache_block {
    int m, n;
    void *cpu_ptr, *gpu_ptr;
};

class GPU_Cache {
    int max_size;
    int total_size;
    std::unordered_map<int, cache_block> cache_map;
    
public:
	GPU_Cache(int size);
    bool get(int m, int n, void **cpu_ptr, void **gpu_ptr);
    bool put(int m, int n, void **cpu_ptr, void **gpu_ptr);
    bool put_and_malloc(int m, int n, void **cpu_ptr, void **gpu_ptr);

private:
	int dim_to_key(int m, int n);

};
