#!/usr/bin/env python3

import os
import sys
import json
import boto3
import utils
from timeit import default_timer as timer


def main():
    # environment variables
    ELASTIC_ENDPOINT = os.environ['ELASTIC_ENDPOINT']
    ELASTIC_INDEX = os.environ['ELASTIC_INDEX']
    S3_ENDPOINT = os.environ['S3_ENDPOINT']
    AWS_REGION = os.environ['AWS_REGION']
    AWS_ACCESS_KEY_ID = os.environ['AWS_ACCESS_KEY_ID']
    AWS_SECRET_ACCESS_KEY = os.environ['AWS_SECRET_ACCESS_KEY']
    BUCKET_NAME = os.environ['BUCKET_NAME']
    STORAGE_PATH = os.environ['STORAGE_PATH']
    LAMBDA_ENDPOINT = os.environ['LAMBDA_ENDPOINT']

    # initialize s3 client
    session = boto3.session.Session()
    s3 = session.client(
        service_name='s3',
        aws_access_key_id=AWS_ACCESS_KEY_ID,
        aws_secret_access_key=AWS_SECRET_ACCESS_KEY,
        endpoint_url=S3_ENDPOINT,
    )

    # test config
    print('\n--- Loading config')
    config_file = sys.argv[1]
    with open(config_file) as json_data:
        config = json.load(json_data)
        utils.pp(config)

    # wait for services to come up
    print('\n--- Waiting for environment to be up ---')
    print('\n---- 1. Lambda ----')
    utils.waitForLambda(LAMBDA_ENDPOINT)
    print('\n---- 2. ElasticSearch ----')
    utils.waitForElastic(ELASTIC_ENDPOINT)
    print('\n---- 3. S3 (localstack) ----')
    utils.waitForS3(s3)

    # cleanup elastic search indices
    print('\n\n--- Initialize ElasticSearch ---')
    print('\n---- 1. Cleaning indices ----')
    utils.clean_up_existing_elastic_templates(ELASTIC_ENDPOINT)

    # copy files to virtual S3 (localstack)
    print('\n\n--- Initialize storage ---')
    print('\n---- 1. Create bucket ----')
    s3.create_bucket(Bucket=BUCKET_NAME)
    print('\n---- 2. Copy files to bucket ----')
    for file in config['files']:
        print(file)
        local_path = "{0}/{1}".format(STORAGE_PATH, file)
        s3.put_object(Bucket=BUCKET_NAME, Key=file, Body=open(local_path, 'rb'))

    # trigger the lambda for every file and validate elastic indices
    print('\n\n--- Lambda test ---')
    print('\n---- 1. Trigger PUT events ----')
    for file in config['files']:
        print(file)
        utils.executeLambda(BUCKET_NAME, file, AWS_REGION, LAMBDA_ENDPOINT)

    print('\n---- 2. Validate ElasticSearch results ----')
    for file in config['files']:
        print(file)
        utils.check_elastic_record(ELASTIC_ENDPOINT, ELASTIC_INDEX, "sample-bucket|{0}".format(file), config['palette'])

try:
    t_start = timer()
    main()
    print('All tests passed')
    sys.exit(0)
except Exception as e:
    print('Test failed: {0}'.format(e))
    sys.exit(1)
finally:
    t_end = timer()
    t_taken = t_end - t_start
    print('Time taken: {0} sec'.format(t_taken))
