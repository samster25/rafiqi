#ifndef __CLASSIFICATION_H__
#define __CLASSIFICATION_H__

#ifdef __cplusplus
extern "C" {
#endif
  typedef void * c_model;
  typedef void * c_mat;

  void classifier_init();
  c_model model_init(char*, char*, char*, char*);
  c_mat make_mat(char *, size_t);
  const char* model_classify(c_model model, c_mat c_img);
  const char** model_classify_batch(c_model model, c_mat* c_imgs, int num);
  void model_destroy(c_model);
#ifdef __cplusplus
}
#endif

#endif
