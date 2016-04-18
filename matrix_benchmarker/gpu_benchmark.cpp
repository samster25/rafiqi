#include <stdlib.h>
#include <stdio.h>
#include <time.h>
#include <sys/time.h>    
#include <cuda.h>
#include <cuda_runtime.h>
#include <cublas_v2.h>
#include <omp.h>
#include <math.h>
#include <string.h>

#define DEV_NUM 0

struct bench_time {
    int trial_num;
    int m;
    int n;
    struct timeval start;
    struct timeval alloc;
    struct timeval disk_IO;
    struct timeval cuda_alloc;
    struct timeval memcopy;
    struct timeval compute;
    struct timeval memcopy_back;
    struct timeval total;
    double rss;
};

const char* CSV_HEADER = "trial_num,m,n,filesize,alloc,disk_IO,cuda_alloc,gpu_memcopy,compute,memcopy_back,total\n";

void 
cpu_gemv_naive(int m, int n, float *A, float *x, float *y) {
//#pragma omp parallel for 
    for (int block=0; block < m; block++) {
        for (int i = 0; i < n; i++) {
            y[block] += A[block*n+i] * x[i];
        }
    }
}


float *
gen_rand_matrix(int m, int n) {
    int i;
    float *rand_mat = (float *) malloc(m * n * sizeof(float));
    for (i = 0; i < m*n; i++) { 
        rand_mat[i] = -1 + 2*((float)rand()/(float)(RAND_MAX)); //generates float [0,1]
    }
    return rand_mat;
}

void
print_mat(int m, int n, float *A) {
    int i,j;
    for (i = 0; i < m; i++) {
        for (j = 0; j < n; j++) {
            printf("%.2f ", A[i*m + j]);
        }
        printf("\n");
    }
}

double
time_diff(struct timeval *tv1, struct timeval *tv2) {
    return (double) (tv2->tv_usec - tv1->tv_usec) / 1000000 +
         (double) (tv2->tv_sec - tv1->tv_sec);
}

void
write_data_header(FILE *f) {
    fputs(CSV_HEADER, f);
}

void
write_data_entry(FILE *f, struct bench_time *dat) {
    fprintf(f,"%d,%d,%d,", dat->trial_num,dat->m, dat->n);
    fprintf(f,"%lu,",dat->m*dat->n*sizeof(float));
    fprintf(f,"%.10f,", time_diff(&dat->start, &dat->alloc));
    fprintf(f,"%.10f,", time_diff(&dat->alloc, &dat->disk_IO));
    fprintf(f,"%.10f,", time_diff(&dat->disk_IO, &dat->cuda_alloc));
    fprintf(f,"%.10f,", time_diff(&dat->cuda_alloc, &dat->memcopy));
    fprintf(f,"%.10f,", time_diff(&dat->memcopy, &dat->compute));
    fprintf(f,"%.10f,", time_diff(&dat->compute, &dat->memcopy_back));
    fprintf(f,"%.10f\n", time_diff(&dat->start, &dat->total));
}

double
error_sum(float *a, float *b, int n) {
    double error = 0.0;
    int i;
    for (i = 0; i < n; i++) {
        error += ((b[i] - a[i])*(b[i] - a[i]));
    }
    return error;
}

void
benchmark_gpu_gemv_file(int m, int n, char *filename, float *x, float *y, struct bench_time *dat) {
    dat->m = m;
    dat->n = n;
    cublasHandle_t handle;
    float al = 1.0f;
    float bet = 0.0f;
    float *A, *cuda_A, *cuda_x, *cuda_y;
    gettimeofday(&dat->start, NULL); 
    
    A = (float *) malloc(m*n*sizeof(float));
    gettimeofday(&dat->alloc, NULL); 
    
    FILE *f = fopen(filename, "r");
    if (fread((void *) A, m*n*sizeof(float), 1,f) == 0)
        return;
    fclose(f);
    gettimeofday(&dat->disk_IO, NULL); 
    
    if ((cudaMalloc((void **) &cuda_A, m*n*sizeof(float)) != cudaSuccess) ||
        (cudaMalloc((void **) &cuda_x, n*sizeof(float)) != cudaSuccess) ||
        (cudaMalloc((void **) &cuda_y, m*sizeof(float)) != cudaSuccess)) {
        printf("error cuda mallocing\n");
        exit(1);
    }
    cublasCreate(&handle); 
    gettimeofday(&dat->cuda_alloc, NULL); 
    
    if (cudaMemcpy(cuda_x, x, n*sizeof(float), cudaMemcpyHostToDevice) != cudaSuccess) {
        printf("cudaMemcpy error\n");
        exit(1);
    }
    cudaMemcpy(cuda_A, A, m*n*sizeof(float), cudaMemcpyHostToDevice);
    gettimeofday(&dat->memcopy, NULL); 
    
    cublasSgemv(handle,CUBLAS_OP_N,m,n,&al, cuda_A,m,cuda_x,1,&bet, cuda_y,1);
    cudaDeviceSynchronize(); 
    gettimeofday(&dat->compute, NULL); 
    
    cudaMemcpy(y, cuda_y, m*sizeof(float), cudaMemcpyDeviceToHost);
    gettimeofday(&dat->memcopy_back, NULL); 
    
    cudaFree(cuda_A);
    cudaFree(cuda_x);
    cudaFree(cuda_y);
    cublasDestroy(handle);
    gettimeofday(&dat->total, NULL); 
    //float *cpu_y = (float *) malloc(sizeof(float) * m);
    //cpu_gemv_naive(m,n,A,x,cpu_y);
    //dat->rss = error_sum(y, cpu_y, m*n);
    //free(cpu_y);
}



int
main(int argc, char **argv) {
    if (argc < 5) {
        printf("Usage: %s matrix_directory manifesto_path output_name num_trials\n", argv[0]);
	    return -1;
    }
    int num_trials = atoi(argv[4]);

    FILE *f = fopen(argv[2], "r");
    if (!f) {
        printf("couldn't open manifest of files: %s\n", argv[2]);
        return -1;
    }
    cudaSetDevice(DEV_NUM);
    char buf[1024];
    char name_buf[1024];
    struct bench_time dat;  
    FILE *out = fopen(argv[3], "w+");
    write_data_header(out);
    printf("Starting Benchmark!\n\n");
    for (int i=0; i < num_trials;i++) {
        printf("Beginning trial #%d\n", i);
        while (fgets (buf, 1024, f)!=NULL) {
            char *p = buf;
            int m = atoi(buf);
            while (*p != 'x' and *p != 0)
                p++;
            p++;
            int n = atoi(p);
            while (*p != '\n' and *p != 0)
                p++;
            *p = 0;
            float *vec;
            if (!(vec = gen_rand_matrix(n,1))) {
                printf("error! gen vector\n");
            }

            float *y;
            if (!(y = (float *)  malloc(m*sizeof(float)))) {
                printf("error mallocing y\n");
            }

            strncpy(name_buf,argv[1],1024);
            strcat(name_buf, "/");
            strcat(name_buf,buf);
            benchmark_gpu_gemv_file(m,n,name_buf,vec, y, &dat);
            dat.trial_num = i;
            write_data_entry(out,&dat);
            printf("  size: %dx%d total time: %f\n", m,n,time_diff(&dat.start, &dat.total));
            free(vec);
            free(y);
        }
        rewind(f);
    }
    fclose(f);
    fclose(out);
    printf("stats date is written to: %s\n", argv[3]);
    printf("done.\n");
}

