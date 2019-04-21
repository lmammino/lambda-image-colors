from time import sleep
import pprint
import subprocess
import lambda_events
import re


def pp(el):
    p = pprint.PrettyPrinter(indent=2)
    p.pprint(el)


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

    raise TimeoutError(
        "S3 did not bootstrap in time ({0} sec)".format(max_wait))


def waitForLambda(host, max_wait=180):
    current_wait_exp = 0
    current_wait = 1
    while (current_wait < max_wait):
        try:
            result = subprocess.Popen(["invok", "--ping", "--host", host])
            result.communicate()
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

    raise TimeoutError(
        "Lambda did not bootstrap in time ({0} sec)".format(max_wait))


def executeLambda(bucket, key, region, endpoint):
    event = lambda_events.create_s3_put(
        {'bucket': bucket, 'key': key, 'region': region})
    proc = subprocess.Popen(["invok", "--host", endpoint],
                            stdout=subprocess.PIPE,
                            stdin=subprocess.PIPE,
                            stderr=subprocess.PIPE)
    out, err = proc.communicate(event.encode())
    if (proc.returncode > 0):
        print('Lambda stderr: {0}'.format(err.decode()))
        raise ValueError('Lambda execution failed')


def validateTags(s3, bucket, key, palette):
    response = s3.get_object_tagging(
        Bucket=bucket,
        Key=key
        )
    for tag in response['TagSet']:
        print(tag['Key'], tag['Value'])
        if re.search(r"^Color[1-4]$", tag['Key']):
            print("{}: {}".format(tag['Key'], tag['Value']))
            assert(tag['Value'] in palette)
