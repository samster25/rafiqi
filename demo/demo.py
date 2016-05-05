import cv2
cap = cv2.VideoCapture(0)
cap.set(cv2.cv.CV_CAP_PROP_FRAME_WIDTH,480)
cap.set(cv2.cv.CV_CAP_PROP_FRAME_HEIGHT,360)
import sys
import json
import httplib
models = ['caffenet','googlenet']

socks = [httplib.HTTPConnection("127.0.0.1", 8000) for each in models]
for each in socks:
    each.connect()
model_socks = zip(models, socks)

i = 0
while True:
    _,img = cap.read()
    cv2.imshow('frame',img)
    if cv2.waitKey(1) & 0xFF == ord('q'):
        break
    data = cv2.imencode('.jpg',img)[1].tobytes()
    requests = [sock.request('POST', '/classify?model_name=' + name, data) for name,sock in model_socks]
    responses = [(name,json.loads(sock.getresponse().read())) for name, sock in model_socks]

    if i == 0:
        sys.stderr.write("\x1b[2J\x1b[H") 
        
        for name, each in responses[0:-1]:
            s = json.dumps(each, sort_keys=True, indent=2, \
                        separators=(',', ': '))
            sys.stdout.write("{0}:\n{1}\n".format(name,s))
        name,each = responses[-1] 
        s = json.dumps(each, sort_keys=True, indent=2, \
                        separators=(',', ': '))
        sys.stdout.write("{0}:\n{1}".format(name,s))
        
        sys.stdout.flush()
    i = (i + 1) % 4
