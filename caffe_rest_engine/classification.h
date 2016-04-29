#ifndef CLASSIFICATION_H
#define CLASSIFICATION_H

#ifdef __cplusplus
extern "C" {
#endif

typedef void * c_model;

void classifier_init();
c_model model_init(char*, char*, char*, char*);
const char* model_classify(c_model, char*, size_t);
void model_destroy(c_model);

#ifdef __cplusplus
}
#endif
#endif
