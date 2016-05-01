time curl -XPOST --data-binary @/work/caffe/examples/images/cat.jpg localhost:8000/classify?model_name=caffenet -s -o /dev/null
