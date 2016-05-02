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
#include <sys/time.h>
#include "classifier.h"
#include "classification.h"
#include <cuda_profiler_api.h>
using namespace caffe;  // NOLINT(build/namespaces)
using std::string;

void classifier_init() {
  ::google::InitGoogleLogging("inference_server");
}

c_model model_init(char* model_file, char* trained_file, char* mean_file, char* label_file) {
  return reinterpret_cast<void*>(new Classifier(string(model_file), string(trained_file), \
                                 string(mean_file), string(label_file)));
}

c_mat make_mat(c_model model, char *buffer, size_t length) {
  Classifier *classifier = reinterpret_cast<Classifier*>(model);
  struct img_processor *ip = classifier->Classifier::Preprocess(buffer, length);
  return reinterpret_cast<c_mat>(ip); 
}

void model_destroy(c_model model) {
  delete reinterpret_cast<Classifier*>(model); 
}

const char *model_classify(c_model model, c_mat c_img) {
  const char **out, *rtn;
  out = model_classify_batch(model, &c_img, 1);
  rtn = out[0];
  free(out); 
  return rtn;
}
long timevaldiff(struct timeval *starttime, struct timeval *finishtime)
{
      long msec;
      msec=(finishtime->tv_sec-starttime->tv_sec)*1000;
      msec+=(finishtime->tv_usec-starttime->tv_usec)/1000;
      return msec;
}

const char** model_classify_batch(c_model model, c_mat* c_imgs, int num)
{
  try
  {
    const char **rtn = (const char **) malloc(num*sizeof(char*)); 
    std::vector<std::vector<Prediction> > all_predictions;
    Classifier *classifier = reinterpret_cast<Classifier*>(model);
    struct img_processor **imgs_ptr = reinterpret_cast<struct img_processor**>(c_imgs);
    std::vector<GpuMat*> imgs;
    for (int i = 0; i < num; i++ ) {
        imgs.push_back(&imgs_ptr[i]->sample_normalized);
    }
    struct timeval start, end;
    gettimeofday(&start, NULL);
    all_predictions = classifier->Classify(imgs);
    gettimeofday(&end, NULL);
    std::cout << "classify ms: " << timevaldiff(&start, &end) << std::endl; 
    /* Write the top N predictions in JSON format. */
    for (int j=0; j < num; j++) {
      std::vector<Prediction> predictions = all_predictions[j];
      std::ostringstream os;
      os << "[";
      for (size_t i = 0; i < predictions.size(); ++i)
      {
        Prediction p = predictions[i];
        os << "{\"confidence\":" << std::fixed << std::setprecision(4)
           << p.second << ",";
        os << "\"label\":" << "\"" << p.first << "\"" << "}";
        if (i != predictions.size() - 1)
          os << ",";
      }
      os << "]";
      errno = 0;
      std::string str = os.str();
      rtn[j] = strdup(str.c_str());
      delete imgs_ptr[j];  
    }
    return rtn;
  }
  catch (const std::invalid_argument&)
  {
    errno = EINVAL;
    return NULL;
  }
}

