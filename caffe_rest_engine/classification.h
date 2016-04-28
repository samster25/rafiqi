#ifndef CLASSIFICATION_H
#define CLASSIFICATION_H

#ifdef __cplusplus
extern "C" {
#endif

typedef void * c_classifier;
c_classifier classifier_initialize(char*, char*, char* mean_file, char*);
const char* classifier_classify(c_classifier, char*, size_t);

#ifdef __cplusplus
}
#endif
#endif
