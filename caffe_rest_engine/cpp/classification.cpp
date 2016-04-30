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

using namespace caffe;  // NOLINT(build/namespaces)
using std::string;

void classifier_init() {
  ::google::InitGoogleLogging("inference_server");
}

c_model model_init(char* model_file, char* trained_file, char* mean_file, char* label_file) {
  return reinterpret_cast<void*>(new Classifier(string(model_file), string(trained_file), \
                                 string(mean_file), string(label_file)));
}

c_mat make_mat(char *buffer, size_t length) {
  cv::_InputArray array(buffer, length);
  cv::Mat img = imdecode(array, -1);
  if (img.empty()) {
    return NULL;
  }
  return reinterpret_cast<c_mat>(new cv::Mat(img)); 
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

const char** model_classify_batch(c_model model, c_mat* c_imgs, int num)
{
  try
  {
    const char **rtn = (const char **) malloc(num*sizeof(char*)); 
    std::vector<std::vector<Prediction> > all_predictions;
    Classifier *classifier = reinterpret_cast<Classifier*>(model);
    cv::Mat **imgs_ptr = reinterpret_cast<cv::Mat**>(c_imgs);
    std::vector<cv::Mat*> imgs(imgs_ptr, imgs_ptr + num); 
    all_predictions = classifier->Classify(imgs);

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
      delete imgs[j];  
    }
      return rtn;
  }
  catch (const std::invalid_argument&)
  {
    errno = EINVAL;
    return NULL;
  }
}

