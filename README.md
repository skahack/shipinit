# shipinit

A secret management tool for container, builds on EC2 SSM Parameter Store.

## Usage

The envload command is a exporting env vars from SSM Parameter Store.

```
$ shipinit envload [flags]

Flags:
      --env string            Environment Name (default "prd")
      --service-name string   Service Name

Example:
      $ shipinit envload --env prd --service-name foo
      > export FOO=bar
      > export BAR=baz
```

### Setting env vars

```ruby

env = "prd"
name = "foo"

# set a KMS ARN
key_id = "arn:aws:kms:<REGION>:<ACCOUNT_ID>:key/<ID>"

map = {
  FOO: "bar",
  BAR: "baz",
}


def command(env, name, key_id, k, v)
  "aws ssm put-parameter --name #{env}.#{name}.#{k} --value \"#{v}\" --type SecureString --key-id #{key_id}"
end

map.each do |k,v|
  system(command(env, name, key_id, k, v))
end
```

### Related articles

- [Managing Secrets for Amazon ECS Applications Using Parameter Store and IAM Roles for Tasks](https://aws.amazon.com/blogs/compute/managing-secrets-for-amazon-ecs-applications-using-parameter-store-and-iam-roles-for-tasks/)

## License

MIT
