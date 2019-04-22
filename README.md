# lambda-image-colors

An example AWS Lambda written in GoLang to tag a picture with its prominent colors

[![buddy pipeline](https://app.buddy.works/lucianomammino/lambda-image-colors/pipelines/pipeline/184112/badge.svg?token=c36f5f6c44fbf89b0e46f07e81533fc6015dd8a58c666de0e8c2c7c9e4bc73c3 "buddy pipeline")](https://app.buddy.works/lucianomammino/lambda-image-colors/pipelines/pipeline/184112)

## About

This repository implements an AWS Lambda using Go that allows you to tag JPEG images in an S3 bucket with their prominent colors from a given palette.

![Lambda trigger schema](/images/lambda-trigger.png)

The goal of this project is to act as a tutorial to learn how to build, test and deploy AWS Lambdas written in Go.

This work is kindly sponsored by [Buddy.works](https://buddy.works).

## Getting started

### Requirements

In order to run this example you need:

- Go (1.12+)
- Terraform (0.11+)
- Docker (18.09+)
- Docker-compose (1.23+)
- GNU make
- An AWS account (with AWS cli installed and configured)

### Folders

- Lambda source code can be found in [`cmd/image-colors-lambda`](/cmd/image-colors-lambda)
- Terraform code for stack definition can be found in [`stack`](/stack)
- Integration tests (using `docker-compose`) can be found in [`tests`](/tests)

## Test and build

To test the application you can run:

```bash
make test
```

This will run:

- Linting checks
- Unit tests
- Integration tests

## Deployment

In order to deploy this solution to your default AWS account you can run:

```bash
make deploy
```

## Read the full tutorial

Yet to be published, stay tuned!

## Contributing

Everyone is very welcome to contribute to this project.
You can contribute just by submitting bugs or suggesting improvements by
[opening an issue on GitHub](https://github.com/lmammino/lambda-image-colors/issues).

## License

Licensed under [MIT License](LICENSE). © Luciano Mammino.
