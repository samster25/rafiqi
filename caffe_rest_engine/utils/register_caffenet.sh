if [ "$#" -ne 1  ]; then
    echo "Usage: $0 path/to/caffe"
fi

./register_model.py caffenet https://raw.githubusercontent.com/jimgoo/caffe-oxford102/master/AlexNet/deploy.prototxt http://dl.caffe.berkeleyvision.org/bvlc_reference_caffenet.caffemodel $1/data/ilsvrc12/imagenet_mean.binaryproto $1/data/ilsvrc12/synsets.txt
