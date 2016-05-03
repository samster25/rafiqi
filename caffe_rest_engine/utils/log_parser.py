import re
import sys

def get_entries(regex, debug_file="../logs/debug.log"):
    results = []
    with open(debug_file) as f:
        for line in f:
            result = re.search(regex, line)
            if result is not None:
                results.append(result.groups())
    return results

def get_request_times(debug_file="../logs/debug.log"):
    reg = r'Request returning success took.*?([0-9]+)ms'
    entries = get_entries(reg, debug_file)
    return map(lambda i: int(i[0]), entries)

def get_classify_times(debug_file):
    reg = r'classify ms.*?([0-9]+)'
    entries = get_entries(reg, debug_file)
    return map(lambda i: int(i[0]), entries)

if __name__ == "__main__":
    print get_classify_times(sys.argv[1])
