for N in {1,2,4,8,16,32,64};
do
    for q in {3,5,7,10};
    do
       curl http://localhost:8000/change_params\?quanta=$q\&batchSize=$N;
       ~/.go/bin/boom -c $[$N*4] -n 4096 -m POST -d @2.jpg http://localhost:8000/classify?model_name=caffenet | tee \
       /scratch/sammy/rafiqi/boom_data/4_$N--$q.csv;
    done
done
