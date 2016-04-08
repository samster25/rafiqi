import numpy as np
import os, sys

dir_name = "/scratch/sammy/matrix_models"

def gen_matrix(m, n, directory):
    data = np.random.randn(m,n).astype(np.float32)
    data.tofile(directory + "/" + str(m) + "x" + str(n))
def main():
    sizes = [2 ** i for i in range(8,15)]
    if not os.path.exists(sys.argv[1]):
            os.makedirs(sys.argv[1])
    print("Generating matrices in {0}!".format(sys.argv[1]))
    print("Writing manifesto at {0}!".format(sys.argv[2]))
    start = int(sys.argv[3])
    end = int(sys.argv[4])
    step = int(sys.argv[5])
    print("generating matrices from {0} to {1} with a step multiplier of {2}!".format(2**start, 2**end, 2**step))
    f = open(sys.argv[2],'w')
    for i in range(start,end+1,step):
        for j in range(start, end+1, step):
            gen_matrix(2**i,2**j, sys.argv[1])
            print(" Made Matrix with size {0}x{1}".format(2**i,2**j))
            f.write(str(2**i) + "x" + str(2**j) + "\n")
    f.close()
if __name__ == "__main__":
    main()  
