# Packer Resource

This is a Concourse CI resource that builds Packer scripts. Both amazon-ebs and
docker types are supported.

# Resource type
You need to add this as a resource type to use it in your pipeline.
```yaml
resource_types:
- name: packer-resource
  type: docker-image
  source:
    repository: pipelineci/packer-resource
```

## Source Configuration

* `type`: *Required.* Either `docker` or `amazon-ebs`.

### Example


``` yaml
- name: myapp-packer-docker
  type: packer-resource
  source:
    type: docker
```

## Behaviour

### `check`:

### `in`:

### `out`: Build the packer script.

#### Parameters
* `source_ami_path`: *Required for amazon-ebs.* Path to file containing the
  source AMI ID. Provided to the packer environment via `-var
  source_ami=<source_ami>`.
* `build_dir`: *Required.* Directory that contains the packer json and
  resources.
* `packer_json`: *Required.* The filename of the packer script.
* `version_dir`: *Required.* Directory that contains the semver resource. And
  will be provided to the packer environment via `-var version=<version>`.
* `aws_access_key_id`: *Optional.* This variables will be provided the packer
  environment via `-var aws_access_key_id=<aws_access_key_id>`.
* `aws_secret_access_key`: *Optional.* This variables will be provided the packer
  environment via `-var aws_secret_access_key=<aws_secret_access_key>`.

