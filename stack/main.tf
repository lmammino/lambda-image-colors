provider "aws" {
  region = "us-east-1"
}

data "aws_region" "selected" {}

data "aws_caller_identity" "selected" {}

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

resource "aws_iam_role_policy" "image-colors-lambda-s3-access" {
  name = "image-colors-lambda-s3-access"
  role = "${aws_iam_role.image-colors-lambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket"
      ],
      "Resource": [
        "${aws_s3_bucket.images_bucket.arn}"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObjectTagging"
      ],
      "Resource": [
        "${aws_s3_bucket.images_bucket.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "image-colors-lambda-cloudwatch-access" {
  name = "image-colors-lambda-cloudwatch-access"
  role = "${aws_iam_role.image-colors-lambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream"
      ],
      "Resource": [
        "arn:aws:logs:${data.aws_region.selected.name}:${data.aws_caller_identity.selected.account_id}:log-group:/aws/lambda/image-colors:*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:PutLogEvents"
      ],
      "Resource": [
        "arn:aws:logs:${data.aws_region.selected.name}:${data.aws_caller_identity.selected.account_id}:log-group:/aws/lambda/image-colors:*:*"
      ]
    }
  ]
}
EOF
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

output "stack_id" {
  value = "${random_id.stack_id.hex}"
}
