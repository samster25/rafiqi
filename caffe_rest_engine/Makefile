CC=g++

CAFFE = /scratch/sammy/caffe
OPENCV = /scratch/sammy/opencv
INCLUDE = -I$(CAFFE)/include \
                    -I$(CAFFE)/src \
                    -I$(CAFFE) \
                    -I/usr/local/cuda/include \
                    -I$(OPENCV)/include

LIBRARY =           -L$(shell pwd)/bin -lclass \
                    -L$(OPENCV)/lib -lopencv_core -lopencv_imgproc -lopencv_imgcodecs \
                    -L/usr/local/x86_64-linux-gnu/ -lprotobuf \
                    -L/usr/lib/x86_64-linux-gnu/ -lglog -lboost_system -lboost_thread \
                    -L/usr/local/cuda/lib64/ -lcudart -lcublas -lcurand \
                    -L$(CAFFE)/build/lib/ -lcaffe \
                    -Wl,-rpath,$(OPENCV)/lib \
                    -Wl,-rpath,$(CAFFE)/build/lib \
                    -Wl,-rpath,$(shell pwd)/bin
server:
	CGO_LDFLAGS='$(LIBRARY)' CGO_CFLAGS='-I$(shell pwd)/cpp' go build -o bin/server_exec go/*.go
compile:
	g++ -c -fpic -DUSE_OPENCV $(INCLUDE) -o ./bin/classification.o ./cpp/classification.cpp
	g++ -shared -o ./bin/libclass.so ./bin/classification.o
	rm ./bin/*.o

