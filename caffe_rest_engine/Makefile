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
CPP_FILES := $(wildcard ./cpp/*.cpp)
OBJ_FILES := $(addprefix ./bin/,$(notdir $(CPP_FILES:.cpp=.o)))

bin/%.o: cpp/%.cpp
	$(CC) $(INCLUDE) -std=c++11 -c -fpic -o $@ $<

compile: $(OBJ_FILES)
	mkdir -p bin
	$(CC) -shared -o ./bin/libclass.so $(OBJ_FILES) 
	rm ./bin/*.o
server:
	CGO_LDFLAGS='$(LIBRARY)' CGO_CFLAGS='-I$(shell pwd)/cpp' go build -o bin/server_exec go/*.go
all: compile server
clean:
	rm -rf ./bin/*  
