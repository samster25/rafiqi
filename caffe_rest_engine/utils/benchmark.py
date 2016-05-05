import os
import sys
import time
import subprocess

maxBatch = [1, 2, 4, 8, 16, 32, 64]
numCtx = [4]
quanta = [1, 2, 3, 4, 5, 7, 10]

def generateCommands():
    cmds = []
    
    def genCmd(batch, ctx, quanta):
        return "bin/server_exec -debug -maxBatch {0} -numContexts {1} -quanta {2}".format(batch, ctx, quanta)
    
    for batchSize in maxBatch:
        for ctx in numCtx:
            for curr_q in quanta:
                cmds.append((genCmd(batchSize, ctx, curr_q), "data_{0}_{1}_{2}.csv".format(batchSize, ctx, curr_q)))
    return cmds

def boomCmd(concurrency, num_reqs, csvname, imagePath, url, model_name):
    return "boom -c {0} -n {1} -m POST -o csv -d @{2} {3}/classify?model_name={4} > {5}.csv".format(concurrency, num_reqs, imagePath, url, model_name, csvname)

if __name__ == "__main__":
    cmds = generateCommands()
    boom_cmds = []
    for cmd in cmds:
        boom_cmds.append(boomCmd(sys.argv[1], sys.argv[2], cmd[1], sys.argv[3], sys.argv[4], sys.argv[5]))
    for i in range(len(cmds)):
        subprocess.Popen(cmds[i][0], shell=True)
        time.sleep(20)
        subprocess.Popen(boom_cmds[i], shell=True)

