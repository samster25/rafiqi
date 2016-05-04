#ifndef CLASSIFICATION_H
#define CLASSIFICATION_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stddef.h>

typedef struct classifier_ctx classifier_ctx;
typedef classifier_ctx* c_model;

void classifier_init();
c_model model_init(char* model_file, char* trained_file,
                                      char* mean_file, char* label_file);

const char* model_classify(c_model model,
                                char* buffer, size_t length);
const char** model_classify_batch(c_model model,
                                char** buffer, size_t *length, size_t num);

void model_destroy(c_model model);

#ifdef __cplusplus
}
#endif

#endif // CLASSIFICATION_H
