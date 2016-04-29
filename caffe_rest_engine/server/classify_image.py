import base64
import json
import requests
import sys


ENDPOINT = "http://localhost:8000/classify"

def main():
    model_name = sys.argv[1]
    image_path = sys.argv[2]

    with open(image_path, 'rb') as f:
        img_raw = f.read()

    b64img = base64.b64encode(img_raw)

    req = {
            'Model': model_name,
            'Image': b64img
    }

    print requests.post(ENDPOINT, json.dumps(req)).text


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: %s model_name <path to image>" % sys.argv[0])

    main()


