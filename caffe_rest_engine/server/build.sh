g++ -c classification.cpp -L../../../caffe/build/lib -lcaffe -lglog -lboost_system -lboost_thread -std=c++11 -I../../../caffe/include -I.. -O2 -fomit-frame-pointer -Wall
ar cru libclassification.a classification.o

