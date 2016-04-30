#!/usr/bin/env python2
import argparse
import base64
import json
import requests
import sys


ENDPOINT = "http://localhost:8000/classify"

def main(model_name, image_path):

    with open(image_path, 'rb') as f:
        img_raw = f.read()

    b64img = base64.b64encode(img_raw)

    req = {
            'Model': model_name,
            'Image': str(b64img)
    }

    print(requests.post(ENDPOINT, json.dumps(req)).text)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("model_name")
    parser.add_argument("image_path")

    args = parser.parse_args()

    main(args.model_name, args.image_path)


