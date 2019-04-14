provider "aws" {
  region = "us-east-1"
}

resource "random_id" "stack_id" {
  byte_length = 16
}

variable "stack_name" {
  default = "image-colors"
}

variable "bucket_name" {
  default = "image-colors"
}

resource "aws_s3_bucket" "images_bucket" {
  bucket = "${var.bucket_name}-${random_id.stack_id.hex}"
  acl    = "private"

  tags = {
    Name     = "${var.bucket_name}-${random_id.stack_id.hex}"
    Stack    = "${var.stack_name}"
    Stack_id = "${random_id.stack_id.hex}"
  }
}

resource "aws_iam_role" "image-colors-lambda" {
  name = "image-colors-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "image-colors-lambda-s3-access" {
  name        = "image-colors-lambda-s3-access"
  description = "Allow image-colors to read images from S3"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "0",
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket"
      ],
      "Resource": "arn:aws:s3:::${aws_s3_bucket.images_bucket.arn}"
    },
    {
      "Sid": "1",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::${aws_s3_bucket.images_bucket.arn}/*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "image-colors-lambda-cloudwatch-access" {
  name        = "image-colors-lambda-cloudwatch-access"
  description = "Allow image-colors to write logs to cloudwatch"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "1",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:::*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "image-colors-lambda-policy-attach-1" {
  role       = "${aws_iam_role.image-colors-lambda.name}"
  policy_arn = "${aws_iam_policy.image-colors-lambda-s3-access.arn}"
}

resource "aws_iam_role_policy_attachment" "image-colors-lambda-policy-attach-2" {
  role       = "${aws_iam_role.image-colors-lambda.name}"
  policy_arn = "${aws_iam_policy.image-colors-lambda-cloudwatch-access.arn}"
}

resource "aws_lambda_function" "image-colors" {
  filename      = "../build/image-colors.zip"
  function_name = "image-colors"
  role          = "${aws_iam_role.image-colors-lambda.arn}"
  handler       = "image-colors"

  source_code_hash = "${filebase64sha256("../build/image-colors.zip")}"

  runtime = "go1.x"

  memory_size                    = 256
  timeout                        = 30
  reserved_concurrent_executions = 10
  publish                        = true

  environment {
    variables = {
      ELASTIC_HOSTS = "${join(",", formatlist("https://%s:%s/", aws_elasticsearch_domain.image-colors.*.endpoint, "9200"))}"
      ELASTIC_INDEX = "image_colors"
    }
  }

  tags = {
    Name     = "image-colors"
    Stack    = "${var.stack_name}"
    Stack_id = "${random_id.stack_id.hex}"
  }
}

resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "image-colors-AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.image-colors.arn}"
  principal     = "s3.amazonaws.com"
  source_arn    = "${aws_s3_bucket.images_bucket.arn}"
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = "${aws_s3_bucket.images_bucket.id}"

  lambda_function {
    lambda_function_arn = "${aws_lambda_function.image-colors.arn}"
    events              = ["s3:ObjectCreated:*"]
    filter_suffix       = ".jpg"
  }
}

resource "aws_elasticsearch_domain" "image-colors" {
  domain_name           = "es-${substr(random_id.stack_id.hex, 0, 16)}"
  elasticsearch_version = "6.5"

  cluster_config {
    instance_type  = "t2.small.elasticsearch"
    instance_count = 1
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  tags = {
    Domain   = "es-${substr(random_id.stack_id.hex, 0, 16)}"
    Stack    = "${var.stack_name}"
    Stack_id = "${random_id.stack_id.hex}"
  }
}

output "stack_id" {
  value = "${random_id.stack_id.hex}"
}

output "elastic_urls" {
  value = ["${join(",", formatlist("https://%s:%s/", aws_elasticsearch_domain.image-colors.*.endpoint, "9200"))}"]
}
