#!/usr/bin/env python
import argparse
import numpy as np
import random
import sys

ZIPF_ALPHA = 1.5

def sweeping_dist(models):
    i = 0
    while True:
        yield models[i % len(models)]
        i += 1

def random_dist(models):
    while True:
        yield random.choice(models)

def zipf_dist(models):
    while True:
        choice = np.random.zipf(ZIPF_ALPHA)
        while choice > len(models):
            choice = np.random.zipf(ZIPF_ALPHA)
        yield models[choice - 1]





DISTRIBUTIONS = {
        'sweeping': sweeping_dist,
        'zipf': zipf_dist,
        'random': random_dist
}


def main():
    parser = argparse.ArgumentParser()

    parser.add_argument("-pattern", help="pattern type", choices=DISTRIBUTIONS.keys(), required=True)
    parser.add_argument("-model", action="append", required=True)
    parser.add_argument("-total", type=int, required=True)
    parser.add_argument("-serverHost", default="localhost:8000")
    parser.add_argument("-iterations", default=1, type=int)
    parser.add_argument("-out", required=True)

    args = parser.parse_args()

    dist_func = DISTRIBUTIONS[args.pattern]

    full_url = "http://%s/classify?model_name=" % args.serverHost

    choices = []
    values = set()
    dist_iter = iter(dist_func(args.model))
    for _ in range(args.total):
        value = next(dist_iter)
        choices.append(value)
        values.add(value)

    with open(args.out, 'w') as f:
        f.write('\n'.join((full_url + m for m in choices)))

    total = float(len(choices))
    for value in sorted(values):
        print "Percent for:", value, "=", round(float(choices.count(value)) / total*100),"%"
if __name__ == "__main__":
    main()
