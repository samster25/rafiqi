#!/usr/bin/env python2
import argparse
import base64
import json
import requests
import sys

def make_request(input):
    result = {}
    if not input:
        return result

    if input.startswith("http://") or input.startswith("https://"):
        # Remote URL
        result = {
                'URL': input
        }
    else:
        with open(input, 'rb') as f:
            contents = f.read()

        b64encoded = base64.b64encode(contents)

        result = {
                'blob': str(b64encoded)
        }
    return result


def register_model(name, model_path, weights_path, means_path, labels_path, server_url):
    inner_data = {
                'model': make_request(model_path),
                'weights': make_request(weights_path),
    }

    if means_path:
        inner_data['means'] = make_request(means_path)
    
    if labels_path:
        inner_data['labels'] = make_request(labels_path)


    data = {
            'models': {
                name: inner_data
                }
            }
    full_url = "http://{0}/register".format(server_url)
    print "Registering", name, "at", full_url, "..."
    try:
        req = requests.post(full_url, json.dumps(data))
    except requests.exceptions.RequestException as e:
        print("Error in request: ", e)
        print("Registering model failed.")
    else:
        print "Result: ", req.text

def main():
    parser = argparse.ArgumentParser(description="Register a model with a Rafiqi server.")

    parser.add_argument("model_name", help="Name of the model.")
    parser.add_argument('model_path', type=str, help="Path to model file. Uses a blob if the file is local, or \
            URL registration if the path is remote (i.e. prefixed with http[s]://).")
    parser.add_argument('weights_path', type=str, help="Path to weights file. Uses a blob if the file is local, \
            or URL registration if the path is remote (i.e. prefixed with http[s]://).")
    parser.add_argument('--means_path', type=str, help="Path to means file. Uses a blob if the file is local, \
            or URL registration if the path is remote (i.e. prefixed with http[s]://).", default="")
    parser.add_argument('--labels_path', type=str, help="Path to labels file. Uses a blob if the file is local, \
            or URL registration if the path is remote (i.e. prefixed with http[s]://).", default="")

    parser.add_argument('--server_url', type=str, default="localhost:8000",
            help="Server URL (without http). Defaults to localhost:8000")
    args = parser.parse_args()

    register_model(
            name=args.model_name,
            model_path=args.model_path,
            weights_path=args.weights_path,
            means_path=args.means_path,
            labels_path=args.labels_path,
            server_url=args.server_url,
    )

if __name__ == "__main__":
    main()
