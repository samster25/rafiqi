./register_model.py googlenet https://raw.githubusercontent.com/BVLC/caffe/master/models/bvlc_googlenet/deploy.prototxt http://dl.caffe.berkeleyvision.org/bvlc_googlenet.caffemodel --means_path $1/data/ilsvrc12/imagenet_mean.binaryproto --labels_path $1/data/ilsvrc12/synset_words.txt

