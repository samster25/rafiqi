#ifndef CLASSIFICATION_H
#define CLASSIFICATION_H

#ifdef __cplusplus
extern "C" {
#endif

typedef void * c_model;
typedef void * c_mat;

void classifier_init();
c_model model_init(char*, char*, char*, char*);
c_mat make_mat(char *buffer, size_t length) {
const char* model_classify(c_model, char*, size_t);
const char** model_classify_batch(c_model model, c_mat* c_imgs, int num)
void model_destroy(c_model);

#ifdef __cplusplus
}
#endif
#endif
