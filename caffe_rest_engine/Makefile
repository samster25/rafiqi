CC=g++

CAFFE = /scratch/sammy/caffe
OPENCV = /scratch/sammy/opencv
INCLUDE = -I$(CAFFE)/include \
                    -I$(CAFFE)/src \
                    -I$(CAFFE) \
                    -I/usr/local/cuda/include \
                    -I$(OPENCV)/include

LIBRARY = -L$(OPENCV)/lib -lopencv_core -lopencv_imgproc -lopencv_imgcodecs \
                    -L/usr/local/x86_64-linux-gnu/ -lprotobuf \
                    -L/usr/lib/x86_64-linux-gnu/ -lglog -lboost_system \
                    -L/usr/local/cuda/lib64/ -lcudart -lcublas -lcurand \
                    -L$(CAFFE)/build/lib/ -lcaffe \
                    -Wl,-rpath,$(OPENCV)/lib \
                    -Wl,-rpath,$(CAFFE)/build/lib

all:
	$(CC) $(INCLUDE) classification.cpp -DUSE_OPENCV -o exercise $(LIBRARY)

test:
	./exercise   models/bvlc_reference_caffenet/deploy.prototxt \
    models/bvlc_reference_caffenet/bvlc_reference_caffenet.caffemodel   data/ilsvrc12/imagenet_mean.binaryproto \
    data/ilsvrc12/synset_words.txt   images/cat.jpg

go_server:
	go build -o bin/server_exec ./server/*.go

compile:
	g++ -c -fpic -DUSEOPENCV -rdynamic $(INCLUDE) classification.cpp
	g++ -shared -o libclass.so classification.o
