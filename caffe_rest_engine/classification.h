#ifndef CLASSIFIER_API_H
#define CLASSIFIER_API_H



typedef void * c_classifier;


#ifdef __cplusplus
extern "C" {
#endif
  c_classifier classifier_initialize(char* model_file, char* trained_file, \
                                        char* mean_file, char* label_file);
  const char* classifier_classify(c_classifier ptr, \
                                char* buffer, size_t length);
#ifdef __cplusplus
}
#endif
#endif
