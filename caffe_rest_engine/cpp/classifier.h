#ifndef __CLASSIFIER_H__
#define __CLASSIFIER_H__

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
#include <opencv2/cudaarithm.hpp>
#include <opencv2/cudaimgproc.hpp>
#include <opencv2/cudawarping.hpp>

using namespace caffe;  // NOLINT(build/namespaces)
using std::string;
using GpuMat = cv::cuda::GpuMat;

/* Pair (label, confidence) representing a prediction. */
typedef std::pair<string, float> Prediction;

struct img_processor {
  cv::Mat input;
  GpuMat img;    
  GpuMat sample;
  GpuMat sample_resized;
  GpuMat sample_float;
  GpuMat sample_normalized;
};

class Classifier {
 public:
  Classifier(const string& model_file,
             const string& trained_file,
             const string& mean_file,
             const string& label_file);

  std::vector<std::vector<Prediction> > Classify(const std::vector<GpuMat*>& img, int N = 5);
  struct img_processor *Preprocess(void *, size_t);

 private:
  void SetMean(const string& mean_file);

  std::vector<float> Predict(const cv::Mat& img);
  std::vector<std::vector<float> > Predict(const std::vector<cv::Mat*>& imgs);
  std::vector<std::vector<float> > Predict(const std::vector<GpuMat*>& imgs);

  //void WrapInputLayer(std::vector<cv::Mat>* input_channels, int n);
  void WrapInputLayer(std::vector<GpuMat>* input_channels, int n);

  void Preprocess(const cv::Mat& img,
                  std::vector<cv::Mat>* input_channels);
 
 private:
  shared_ptr<Net<float> > net_;
  cv::Size input_geometry_;
  int num_channels_;
  int output_N_;
  GpuMat mean_;
  std::vector<string> labels_;
};
#endif
