from time import sleep
import traceback
import pprint
import requests
import json
import glob
import os
import subprocess
import lambda_events
from pathlib import Path


def pp(el):
    p = pprint.PrettyPrinter(indent=2)
    p.pprint(el)

def waitForElastic(host, max_wait=120):
    current_wait_exp = 0
    current_wait = 1
    health_check_url = "{0}/?pretty".format(host)
    while (current_wait < max_wait):
        try:
            requests.get(health_check_url)
            return True
        except requests.exceptions.ConnectionError:
            print('\t\t.')
        except Exception as e:
            print('\tIgnored error: {0}'.format(str(e)))

        # exponential fallback
        current_wait_exp += 1
        current_wait = 1 << current_wait_exp
        sleep(current_wait)

    raise TimeoutError("ElasticSearch did not bootstrap in time ({0} sec)".format(max_wait))


def check_elastic_record(host, index, id, palette):
    # makes sure the pending writes are flushed
    flush_url = "{0}/{1}/_flush".format(host, index)
    requests.post(flush_url)

    search_url = "{0}/{1}/default/{2}".format(host, index, id)
    response_raw = requests.get(search_url)
    response = json.loads(response_raw.text)
    assert response['found'], "Record {0} not found".format(id)
    colors = response['_source']['colors']
    assert len(colors) > 0, "Colors array is empty for record {0}".format(id)
    for color in colors:
        assert color in palette, "color {0} found in {1} is not defined in palette".format(color, id)

def clean_up_existing_elastic_templates(host):
    request_url = "{0}/_template/*".format(host)
    response_raw = requests.delete(request_url)
    response = json.loads(response_raw.text)
    assert(response["acknowledged"])


def waitForS3(s3, max_wait=120):
    current_wait_exp = 0
    current_wait = 1
    while (current_wait < max_wait):
        try:
            s3.list_buckets()
            return True
        except Exception as e:
            print("Ignored error: {0}".format(str(e)))

        # exponential fallback
        current_wait_exp += 1
        current_wait = 1 << current_wait_exp
        sleep(current_wait)

    raise TimeoutError("S3 did not bootstrap in time ({0} sec)".format(max_wait))


def waitForLambda(host, max_wait=180):
    current_wait_exp = 0
    current_wait = 1
    while (current_wait < max_wait):
        try:
            result = subprocess.Popen(["invok", "--ping", "--host", host])
            text = result.communicate()[0]
            returncode = result.returncode
            if (returncode > 0):
                raise Exception('Lambda not ready yet')
            return True
        except Exception as e:
            print("Ignored error: {0}".format(str(e)))

        # exponential fallback
        current_wait_exp += 1
        current_wait = 1 << current_wait_exp
        sleep(current_wait)

    raise TimeoutError("Lambda did not bootstrap in time ({0} sec)".format(max_wait))

def executeLambda(bucket, key, region, endpoint):
    event = lambda_events.create_s3_put({ 'bucket': bucket, 'key': key, 'region': region })
    proc = subprocess.Popen(["invok", "--host", endpoint], stdout=subprocess.PIPE, stdin=subprocess.PIPE, stderr=subprocess.PIPE)
    out, err = proc.communicate(event.encode())
    if (proc.returncode > 0):
        print('Lambda stderr: {0}'.format(err.decode()))
        raise ValueError('Lambda execution failed')
