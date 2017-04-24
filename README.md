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

* ``: *Required.*

### Example


``` yaml
```

## Behaviour

### `check`:

### `in`:

### `out`: Build the packer script.

#### Parameters
* ``: *Required.*
* ``: *Optional.*
