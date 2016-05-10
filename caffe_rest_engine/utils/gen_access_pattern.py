#!/usr/bin/env python
import argparse
import numpy as np
import random
import sys

def sweeping_dist(models, total_requests, full_url):
    for i in range(total_requests):
        print full_url  + models[i % len(models)]

def random_dist(models, total_requests, full_url):
    for _ in range(total_requests):
        print full_url + random.choice(models)

def zipf_dist(models, total_requests, full_url):
    for _ in range(total_requests):
        choice = np.random.zipf(1.01)
        while choice > len(models):
            choice = np.random.zipf(1.01)
        print full_url + models[choice - 1]

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

    args = parser.parse_args()

    dist_func = DISTRIBUTIONS[args.pattern]

    full_url = "http://%s/classify?model_name=" % args.serverHost
    for _ in range(args.iterations):
        dist_func(args.model, args.total, full_url)

if __name__ == "__main__":
    main()
