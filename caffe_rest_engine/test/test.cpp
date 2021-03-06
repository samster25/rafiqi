#include <caffe/caffe.hpp>
#include <opencv2/core/core.hpp>
#include <opencv2/highgui/highgui.hpp>
#include <opencv2/imgproc/imgproc.hpp>
#include <algorithm>
#include <iosfwd>
#include <memory>
#include <string>
#include <utility>
#include <vector>
#include "classifier.h"
#include "classification.h"
#include <cuda_profiler_api.h>
using namespace caffe;  // NOLINT(build/namespaces)
using std::string;

/* Pair (label, confidence) representing a prediction. */
typedef std::pair<string, float> Prediction;


int main(int argc, char** argv) {
  if (argc != 6) {
    std::cerr << "Usage: " << argv[0]
              << " deploy.prototxt network.caffemodel"
              << " mean.binaryproto labels.txt img.jpg" << std::endl;
    return 1;
  }
  cudaProfilerStop();
  classifier_init();
  char *model_file   = argv[1];
  char *trained_file = argv[2];
  char *mean_file    = argv[3];
  char *label_file   = argv[4];
  c_model model = model_init(model_file, trained_file, mean_file, label_file);
  std::cout << "built classifer\n"; 
  char *file_name = argv[5];
 
  std::ifstream file;
  file.open(file_name, std::ios::binary);
  std::cout << "opened file " << file_name << "\n";

  file.seekg(0, std::ios::end);
  int size = file.tellg();
  file.seekg(0, std::ios::beg);
  std::cout << "file size: " << size << "\n";
  std::vector<char> buffer(size);
  std::cout << "before read" << file << "\n";
  file.read(buffer.data(),size);
  std::cout << "before classify\n";
  c_mat im = make_mat(model, buffer.data(), size);
  const char *out = model_classify(model, im);
  std::cout << -1 << out << std::endl;
  free((void *) out);
  
  cudaProfilerStart();
  for (int i =0; i < 200; i++) {
    im = make_mat(model, buffer.data(), size);
    out = model_classify(model, im);
    std::cout << i << out << std::endl;
    free((void *) out);
  }
  cudaProfilerStop();

  model_destroy(model);
  /* Print the top N predictions. */
}
