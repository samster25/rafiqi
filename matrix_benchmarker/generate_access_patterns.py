import shutil, os, sys
import numpy as np
import random

def genTotalManifesto(patternGenerator):
	with open(sys.argv[1], 'r') as f:
		availableMats = f.read().strip().split("\n")
		accessPattern = patternGenerator(availableMats)
	with open(sys.argv[2], 'w') as f:
		for el in accessPattern:
			f.write(el + "\n")

def standardPattern(allMats):
	return allMats


def randomSampling(allMats):
	final = []
	for _ in range(len(allMats)*2):
		final.append(random.choice(allMats))
	return final

if __name__ == "__main__":
	if sys.argv[3] == "random":
		genTotalManifesto(randomSampling)
	elif sys.argv[3] == "standard":
		genTotalManifesto(standardPattern)
