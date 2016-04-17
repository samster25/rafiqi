CC=g++

CAFFE = /scratch/sammy/caffe
OPENCV = /usr/local/include/opencv
INCLUDE = -I$(CAFFE)/include \
                    -I$(CAFFE)/src \
                    -I$(CAFFE) \
                    -I/usr/local/cuda/include \
                    -I$(OPENCV)/include

LIBRARY = -L/usr/local/x86_64-linux-gnu/ -lprotobuf \
                    -L/usr/lib/x86_64-linux-gnu/ -lglog -lboost_system \
                    -L/usr/local/cuda/lib64/ -lcudart -lcublas -lcurand \
                    -L/usr/local/lib/ -lopencv_core -lopencv_imgproc -lopencv_highgui \
                    -L$(CAFFE)/build/lib/ -lcaffe

all:
	LD_LIBRARY_PATH=$(CAFFE)/build/lib $(CC) $(INCLUDE) classification.cpp -DUSE_OPENCV -o exercise $(LIBRARY)



