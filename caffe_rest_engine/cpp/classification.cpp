#include "classification.h"

#include <iosfwd>
#include <vector>
#include <stdint.h>

#define USE_CUDNN 1
#include <caffe/caffe.hpp>
#include <caffe/net.hpp>
#include <opencv2/core/cuda.hpp>

#include <opencv2/cudaarithm.hpp>
#include <opencv2/cudaimgproc.hpp>
#include <opencv2/cudawarping.hpp>
#include <opencv2/highgui/highgui.hpp>
#include <opencv2/core/cuda_stream_accessor.hpp>

#include "common.h"
#include "gpu_allocator.h"

using namespace caffe;
using std::string;
using GpuMat = cv::cuda::GpuMat;
using namespace cv;
constexpr static int imageBufferSize = 128 * 1024 * 1024;

/* C bindings cudaMemGetInfo */
uint64_t get_total_gpu_memory() {
  size_t free;
  size_t total;

  cudaMemGetInfo(&free, &total);
  return (uint64_t) total;
}

uint64_t get_free_gpu_memory() {
  size_t free;
  size_t total;

  cudaMemGetInfo(&free, &total);
  return (int64_t) free;
}
/* Pair (label, confidence) representing a prediction. */
typedef std::pair<string, float> Prediction;

/* Based on the cpp_classification example of Caffe, but with GPU
 * image preprocessing and a simple memory pool. */
class Classifier
{
public:
    Classifier(const string& model_file,
               const string& trained_file,
               const string& mean_file,
               const string& label_file,
               size_t max_batch_size);

    Classifier(const string& model_file,
               const string& trained_file,
               const string& mean_file,
               const string& label_file,
               Classifier *old);
    
    GPUAllocator* allocator_;
    std::vector<std::vector<Prediction> > Classify(const std::vector<Mat>& imgs, int N = 5);
    size_t memory_used();
    void move_to_cpu();
    void move_to_gpu();
    void move_to_gpu_async();

private:
    void SetMean(const string& mean_file);

    std::vector<std::vector<float> > Predict(const std::vector<Mat>& imgs);

    void WrapInputLayer(std::vector<GpuMat>* input_channels, int i);

    void Preprocess(const Mat& img,
                    std::vector<GpuMat>* input_channels);
    
private:
    std::shared_ptr<Net<float>> net_;
    Size input_geometry_;
    int num_channels_;
    GpuMat mean_;
    Mat host_mean_;     
    std::vector<string> labels_;
    size_t max_batch_size_;
    cudaStream_t stream_;
    cv::cuda::Stream cvstream_;
};

Classifier::Classifier(const string& model_file,
                       const string& trained_file,
                       const string& mean_file,
                       const string& label_file,
                       size_t max_batch_size)
                        : max_batch_size_(max_batch_size)
{

    Caffe::set_mode(Caffe::GPU);
    allocator_ = new GPUAllocator(imageBufferSize);
    cudaStreamCreate(&stream_);
    cvstream_ = cuda::StreamAccessor::wrapStream(stream_);
    /* Load the network. */
    net_ = std::make_shared<Net<float>>(model_file, TEST);
    net_->CopyTrainedLayersFrom(trained_file);

    CHECK_EQ(net_->num_inputs(), 1) << "Network should have exactly one input.";
    CHECK_EQ(net_->num_outputs(), 1) << "Network should have exactly one output.";

    Blob<float>* input_layer = net_->input_blobs()[0];
    num_channels_ = input_layer->channels();
    CHECK(num_channels_ == 3 || num_channels_ == 1)
        << "Input layer should have 1 or 3 channels.";
    input_geometry_ = Size(input_layer->width(), input_layer->height());

    /* Load the binaryproto mean file. */
    SetMean(mean_file);

    /* Load labels. */
    std::ifstream labels(label_file.c_str());
    CHECK(labels) << "Unable to open labels file " << label_file;
    string line;
    while (std::getline(labels, line))
        labels_.push_back(string(line));

    Blob<float>* output_layer = net_->output_blobs()[0];
    CHECK_EQ(labels_.size(), output_layer->channels())
        << "Number of labels is different from the output layer dimension.";
    input_layer->Reshape(max_batch_size_, num_channels_,
                         input_geometry_.height, input_geometry_.width);
    net_->Reshape();
}

Classifier::Classifier(const string& model_file,
                       const string& trained_file,
                       const string& mean_file,
                       const string& label_file,
                       Classifier *old)
{
    Caffe::set_mode(Caffe::GPU);
    /* Load the network. */
    allocator_ = new GPUAllocator(imageBufferSize);

    net_ = std::make_shared<Net<float>>(model_file, TEST);
    //net_->ShareTrainedLayersWith(old->net_.get());
    
    CHECK_EQ(net_->num_inputs(), 1) << "Network should have exactly one input.";
    CHECK_EQ(net_->num_outputs(), 1) << "Network should have exactly one output.";

    Blob<float>* input_layer = net_->input_blobs()[0];
    num_channels_ = input_layer->channels();
    CHECK(num_channels_ == 3 || num_channels_ == 1)
        << "Input layer should have 1 or 3 channels.";
    input_geometry_ = Size(input_layer->width(), input_layer->height());

    /* Load the binaryproto mean file. */
    //SetMean(mean_file);
    mean_ = old->mean_;
    /* Load labels. */
    //std::ifstream labels(label_file.c_str());
    //CHECK(labels) << "Unable to open labels file " << label_file;
    //string line;
    //while (std::getline(labels, line))
    //    labels_.push_back(string(line));
    labels_ = old->labels_;
    Blob<float>* output_layer = net_->output_blobs()[0];
    CHECK_EQ(labels_.size(), output_layer->channels())
        << "Number of labels is different from the output layer dimension.";

    input_layer->Reshape(1, num_channels_,
                         input_geometry_.height, input_geometry_.width);
    net_->Reshape();

}

static bool PairCompare(const std::pair<float, int>& lhs,
                        const std::pair<float, int>& rhs)
{
    return lhs.first > rhs.first;
}

/* Return the indices of the top N values of vector v. */
static std::vector<int> Argmax(const std::vector<float>& v, int N)
{
    std::vector<std::pair<float, int>> pairs;
    for (size_t i = 0; i < v.size(); ++i)
        pairs.push_back(std::make_pair(v[i], i));
    std::partial_sort(pairs.begin(), pairs.begin() + N, pairs.end(), PairCompare);

    std::vector<int> result;
    for (int i = 0; i < N; ++i)
        result.push_back(pairs[i].second);
    return result;
}

std::vector<std::vector<Prediction> > Classifier::Classify(const std::vector<Mat>& imgs, int N)
{
    std::vector<std::vector<float> > outputs = Predict(imgs);
    std::vector<std::vector<Prediction> > all_predictions;
    N = std::min<int>(outputs[0].size(), N);
    for (int k = 0; k < imgs.size(); k++) {
        std::vector<Prediction> predictions;
        std::vector<float> output = outputs[k];
        std::vector<int> maxN = Argmax(output, N);
        
        for (int i = 0; i < N; ++i)
        {
            int idx = maxN[i];
            predictions.push_back(std::make_pair(labels_[idx], output[idx]));
        }
        all_predictions.push_back(predictions);  
    }
    return all_predictions;
}

/* Load the mean file in binaryproto format. */
void Classifier::SetMean(const string& mean_file)
{
    BlobProto blob_proto;
    ReadProtoFromBinaryFileOrDie(mean_file.c_str(), &blob_proto);

    /* Convert from BlobProto to Blob<float> */
    Blob<float> mean_blob;
    mean_blob.FromProto(blob_proto);
    CHECK_EQ(mean_blob.channels(), num_channels_)
        << "Number of channels of mean file doesn't match input layer.";

    /* The format of the mean file is planar 32-bit float BGR or grayscale. */
    std::vector<Mat> channels;
    float* data = mean_blob.mutable_cpu_data();
    for (int i = 0; i < num_channels_; ++i)
    {
        /* Extract an individual channel. */
        Mat channel(mean_blob.height(), mean_blob.width(), CV_32FC1, data);
        channels.push_back(channel);
        data += mean_blob.height() * mean_blob.width();
    }

    /* Merge the separate channels into a single image. */
    Mat packed_mean;
    merge(channels, packed_mean);

    /* Compute the global mean pixel value and create a mean image
     * filled with this value. */
    Scalar channel_mean = mean(packed_mean);
    host_mean_ = Mat(input_geometry_, packed_mean.type(), channel_mean);
    mean_.upload(host_mean_);
}

std::vector<std::vector<float> > Classifier::Predict(const std::vector<Mat>& imgs)
{
    /* Forward dimension change to all layers. */
    std::vector<GpuMat> float_norm_imgs;
    for (int i = 0; i < imgs.size(); i++) {
      Preprocess(imgs[i], &float_norm_imgs);
    }
    cudaStreamSynchronize(stream_);
    
    Blob<float>* input_layer = net_->input_blobs()[0];
    input_layer->Reshape(imgs.size(), num_channels_,
                         input_geometry_.height, input_geometry_.width);
    net_->Reshape();
    
    for (int i = 0; i < imgs.size(); i++) {
      std::vector<GpuMat> input_channels;
      WrapInputLayer(&input_channels, i);
      cuda::split(float_norm_imgs[i], input_channels);
    }
    net_->Forward();



    /* Copy the output layer to a std::vector */
    Blob<float>* output_layer = net_->output_blobs()[0];
    std::vector<std::vector<float> > outputs;
    for (int i = 0; i < output_layer->num(); ++i) {
      const float* begin = output_layer->cpu_data() + i * output_layer->channels();
      const float* end = begin + output_layer->channels();
      outputs.push_back(std::vector<float>(begin, end));
    }
    return outputs;
}

void Classifier::WrapInputLayer(std::vector<GpuMat>* input_channels, int i)
{
    Blob<float>* input_layer = net_->input_blobs()[0];

    int width = input_layer->width();
    int height = input_layer->height();
    int channels = input_layer->channels();
    float* input_data = input_layer->mutable_gpu_data() + i * width * height * channels;
    for (int i = 0; i < channels; ++i)
    {
        GpuMat channel(height, width, CV_32FC1, input_data);
        input_channels->push_back(channel);
        input_data += width * height;
    }
}

void Classifier::Preprocess(const Mat& host_img,
                            std::vector<GpuMat>* float_norm_imgs)
{
    GpuMat img(host_img, allocator_);
    /* Convert the input image to the input image format of the network. */
    GpuMat sample(allocator_);
    if (img.channels() == 3 && num_channels_ == 1)
        cuda::cvtColor(img, sample, CV_BGR2GRAY, 0, cvstream_);
    else if (img.channels() == 4 && num_channels_ == 1)
        cuda::cvtColor(img, sample, CV_BGRA2GRAY, 0, cvstream_);
    else if (img.channels() == 4 && num_channels_ == 3)
        cuda::cvtColor(img, sample, CV_BGRA2BGR, 0, cvstream_);
    else if (img.channels() == 1 && num_channels_ == 3)
        cuda::cvtColor(img, sample, CV_GRAY2BGR, 0, cvstream_);
    else
        sample = img;

    GpuMat sample_resized(allocator_);
    if (sample.size() != input_geometry_)
        cuda::resize(sample, sample_resized, input_geometry_,0,0,INTER_LINEAR, cvstream_);
    else
        sample_resized = sample;

    GpuMat sample_float(allocator_);
    if (num_channels_ == 3)
        sample_resized.convertTo(sample_float, CV_32FC3, cvstream_);
    else
        sample_resized.convertTo(sample_float, CV_32FC1, cvstream_);

    GpuMat sample_normalized(allocator_);
    cuda::subtract(sample_float, mean_, sample_normalized, noArray(),-1,cvstream_);
    float_norm_imgs->push_back(sample_normalized); 
}

size_t Classifier::memory_used() {
    return net_->memory_used();

}
void Classifier::move_to_cpu() {
    net_->moveToCPU();
    //mean_.refcount = 0;
    //mean_.release();
    delete allocator_;
}

void Classifier::move_to_gpu_async() {
    allocator_ = new GPUAllocator(imageBufferSize);
    //mean_.upload(host_mean_);
    net_->moveToGPUAsync(stream_);

}

void Classifier::move_to_gpu() {
    allocator_ = new GPUAllocator(imageBufferSize);
    //mean_.upload(host_mean_);
    net_->moveToGPU();

}


/* By using Go as the HTTP server, we have potentially more CPU threads than
 * available GPUs and more threads can be added on the fly by the Go
 * runtime. Therefore we cannot pin the CPU threads to specific GPUs.  Instead,
 * when a CPU thread is ready for inference it will try to retrieve an
 * execution context from a queue of available GPU contexts and then do a
 * cudaSetDevice() to prepare for execution. Multiple contexts can be allocated
 * per GPU. */
class CaffeContext
{
public:
    friend ScopedContext<CaffeContext>;

    static bool IsCompatible(int device)
    {
        cudaError_t st = cudaSetDevice(device);
        if (st != cudaSuccess)
            return false;

        cuda::DeviceInfo info;
        if (!info.isCompatible())
            return false;

        return true;
    }

    CaffeContext(const string& model_file,
                 const string& trained_file,
                 const string& mean_file,
                 const string& label_file,
                 int device,
                 size_t max_batch_size)
        : device_(device)
    {
        cudaError_t st = cudaSetDevice(device_);
        if (st != cudaSuccess)
            throw std::invalid_argument("could not set CUDA device");

        //allocator_.reset(new GPUAllocator(1024 * 1024 * 128));
        caffe_context_.reset(new Caffe);
        Caffe::Set(caffe_context_.get());
        classifier_.reset(new Classifier(model_file, trained_file,
                                         mean_file, label_file, max_batch_size));
        Caffe::Set(nullptr);
    }
    
    CaffeContext(const string& model_file,
                 const string& trained_file,
                 const string& mean_file,
                 const string& label_file,
                 Classifier *old,
                 int device)
        : device_(device)
    {
        cudaError_t st = cudaSetDevice(device_);
        if (st != cudaSuccess)
            throw std::invalid_argument("could not set CUDA device");

        //allocator_.reset(new GPUAllocator(1024 * 1024 * 128));
        caffe_context_.reset(new Caffe);
        Caffe::Set(caffe_context_.get());
        classifier_.reset(new Classifier(model_file, trained_file,
                                         mean_file, label_file,
                                         old));
        Caffe::Set(nullptr);
    }

    Classifier* CaffeClassifier()
    {
        return classifier_.get();
    }

private:
    void Activate()
    {
        cudaError_t st = cudaSetDevice(device_);
        if (st != cudaSuccess)
            throw std::invalid_argument("could not set CUDA device");
        classifier_->allocator_->reset();
        Caffe::Set(caffe_context_.get());
    }

    void Deactivate()
    {
        Caffe::Set(nullptr);
    }

private:
    int device_;
    //std::unique_ptr<GPUAllocator> allocator_;
    std::unique_ptr<Caffe> caffe_context_;
    std::unique_ptr<Classifier> classifier_;
};


void classifier_init() {
  ::google::InitGoogleLogging("inference_server");
}


struct classifier_ctx
{
    ContextPool<CaffeContext> pool;
    size_t k;
};

/* Currently, 2 execution contexts are created per GPU. In other words, 2
 * inference tasks can execute in parallel on the same GPU. This helps improve
 * GPU utilization since some kernel operations of inference will not fully use
 * the GPU. */

c_model model_init(char* model_file, char* trained_file,
                                      char* mean_file, char* label_file, size_t num_contexts, size_t max_batch_size)
{
    try
    {
        classifier_ctx* ctx = new classifier_ctx; 
        int device_count; 
        cudaError_t st = cudaGetDeviceCount(&device_count);
        if (st != cudaSuccess)
            throw std::invalid_argument("could not list CUDA devices");
        device_count = 1;
        ContextPool<CaffeContext> pool;
        ctx->k = num_contexts; 
        for (int dev = 0; dev < device_count; ++dev)
        {
            if (!CaffeContext::IsCompatible(dev))
            {
                LOG(ERROR) << "Skipping device: " << dev;
                continue;
            }
            std::cout << "Init'ing Device: " <<  dev << std::endl;
            for (int i = 0; i < num_contexts; ++i)
            {
                std::unique_ptr<CaffeContext> shared_context(new CaffeContext(model_file, trained_file, mean_file,
                                                                       label_file, dev, max_batch_size));
                ctx->pool.Push(std::move(shared_context));
            }
        }
        
        if (ctx->pool.Size() == 0)
            throw std::invalid_argument("no suitable CUDA device");

        /* Successful CUDA calls can set errno. */
        errno = 0;
        return (c_model) ctx;
    }
    catch (const std::invalid_argument& ex)
    {
        LOG(ERROR) << "exception: " << ex.what();
        errno = EINVAL;
        return nullptr;
    }
}

void move_to_cpu(c_model model) {
    classifier_ctx *ctx = (classifier_ctx *) model;
    {
        for (int i = 0; i < ctx->k; i++) {
            ScopedContext<CaffeContext> context(ctx->pool);
            auto classifier = context->CaffeClassifier();
            classifier->move_to_cpu(); 
        }
    }
}
//void move_to_cpu(c_model model) {
//    classifier_ctx *ctx = (classifier_ctx *) model;
//    for (int i = 0; i < ctx->k; i++) {
//        ctx->classifiers[i]->move_to_cpu();
//    }
//}

//void move_to_gpu(c_model model) {
//    classifier_ctx *ctx = (classifier_ctx *) model;
//    for (int i = 0; i < ctx->k; i++) {
//        ctx->classifiers[i]->move_to_gpu();
//    }
//}

void move_to_gpu_async(c_model model) {
    classifier_ctx *ctx = (classifier_ctx *) model;
    {
        for (int i = 0; i < ctx->k; i++) {
            ScopedContext<CaffeContext> context(ctx->pool);
            auto classifier = context->CaffeClassifier();
            classifier->move_to_gpu_async(); 
        }
    }
}
void move_to_gpu(c_model model) {
    classifier_ctx *ctx = (classifier_ctx *) model;
    {
        for (int i = 0; i < ctx->k; i++) {
            ScopedContext<CaffeContext> context(ctx->pool);
            auto classifier = context->CaffeClassifier();
            classifier->move_to_gpu(); 
        }
    }
}


const char** model_classify_batch(c_model model,
                                char** buffer, size_t *length, size_t num)
{
    try
    {
        classifier_ctx *ctx = (classifier_ctx *) model;
        std::vector<Mat> imgs;
        
        for (int i = 0; i < num; i++) { 
            _InputArray array(buffer[i], length[i]);
        
            Mat img = imdecode(array, -1);
            if (img.empty())
                throw std::invalid_argument("could not decode image");
            imgs.push_back(img);
        }
        std::vector<std::vector<Prediction> > all_predictions;
        {
            /* In this scope a Caffe context is acquired for inference and it
             * will be automatically released back to the context pool when
             * exiting this scope. */
            ScopedContext<CaffeContext> context(ctx->pool);
            auto classifier = context->CaffeClassifier();
            all_predictions = classifier->Classify(imgs);
        }
        const char **output = (const char **) malloc(num*sizeof(const char *));
        /* Write the top N predictions in JSON format. */
        for (int j = 0; j < all_predictions.size(); j++) {
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
            output[j] =  strdup(str.c_str());
        }
        return output;
    }
    catch (const std::invalid_argument&)
    {
        errno = EINVAL;
        return nullptr;
    }
}

//size_t get_mem_used(c_model model) {
//    classifier_ctx *ctx = (classifier_ctx *) model;
//    size_t total;
//    for (int i = 0; i < kContextsPerDevice; i++) {
//        std::cout << "mem used: " << ctx->classifiers[i]->memory_used() << std::endl;
//    }
//    return 5;
//}



const char* model_classify(c_model model,
                                char* buffer, size_t length)
{
    classifier_ctx *ctx = (classifier_ctx *) model;
    const char **out =  model_classify_batch(model, &buffer, &length, 1);
    const char *rtn = *out;
    free(out);
    return rtn;
}

void model_destroy(c_model model)
{
    classifier_ctx *ctx = (classifier_ctx *) model;
    {
        for (int i = 0; i < ctx->k; i++) {
            ScopedContext<CaffeContext> context(ctx->pool);
            auto classifier = context->CaffeClassifier();
            delete classifier->allocator_; 
        }
    }
    delete ctx;
}
