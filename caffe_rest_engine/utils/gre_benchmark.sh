for N in {1,2,4,8,16,32,64,128};
do
   ~/.go/bin/boom -c $[$N*4] -n 8192 -m POST -d @2.jpg http://localhost:8000/api/classify | tee \
   /scratch/sammy/rafiqi/gre_data/4_$N.csv;
done
