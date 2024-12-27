# Service Placement Configuration

## Host Placement

You can control which services are deployed to specific hosts using labels in your docker-compose.yml file:

```yaml
services:
  database:
    image: postgres:13
    labels:
      bluetongue.placement.host: "db.example.com"  # Deploy to specific host
      bluetongue.placement.constraints: "disk=ssd"  # Host must have SSD

  cache:
    image: redis:latest
    labels:
      bluetongue.placement.hosts: "cache1.example.com,cache2.example.com"  # Deploy to multiple hosts
      bluetongue.placement.zone: "us-east"  # Deploy only to hosts in us-east zone

  webapp:
    image: nginx:latest
    labels:
      bluetongue.placement.strategy: "spread"  # Spread across available hosts
```

## Deployment Examples

### Deploy to Specific Hosts

```bash
# Deploy to multiple hosts with service placement
deploy --hosts "host1.example.com,host2.example.com,host3.example.com" \
  --ssh-user deploy \
  -i ~/.ssh/prod_key
```

The services will be placed according to their labels:
- `database` will only deploy to `db.example.com`
- `cache` will deploy to both `cache1.example.com` and `cache2.example.com`
- `webapp` will be spread across available hosts

### Host Groups

You can also define host groups in a configuration file:

```yaml:deploy.yml
host_groups:
  database:
    - db1.example.com
    - db2.example.com
    attributes:
      disk: ssd
      role: database

  cache:
    - cache1.example.com
    - cache2.example.com
    attributes:
      role: cache
      zone: us-east

  web:
    - web1.example.com
    - web2.example.com
    - web3.example.com
    attributes:
      role: frontend
```

Then reference the configuration:

```bash
deploy -f docker-compose.yml --config deploy.yml --ssh-user deploy -i ~/.ssh/prod_key
```

## Placement Constraints

### Host Labels
```yaml
services:
  myapp:
    image: myapp:latest
    labels:
      bluetongue.placement.constraints: "role=frontend,zone=us-east"
```

### Resource Requirements
```yaml
services:
  database:
    image: postgres:13
    labels:
      bluetongue.placement.constraints: "disk=ssd,memory>=16gb"
    deploy:
      resources:
        limits:
          memory: 16G
          cpus: '4'
```

### Anti-Affinity Rules
```yaml
services:
  redis-master:
    image: redis:latest
    labels:
      bluetongue.placement.anti-affinity: "redis-slave"  # Don't place on same host as slaves

  redis-slave:
    image: redis:latest
    labels:
      bluetongue.placement.anti-affinity: "redis-master"  # Don't place on same host as master
```

## Placement Strategies

Available placement strategies:

- `spread`: Spread services across available hosts (default)
- `packed`: Pack services onto as few hosts as possible
- `balanced`: Balance services based on resource usage
- `isolated`: Ensure services run on dedicated hosts

```yaml
services:
  webapp:
    image: nginx:latest
    labels:
      bluetongue.placement.strategy: "balanced"
      bluetongue.placement.constraints: "role=frontend"
```

## Host Selection Priority

The deployment tool selects hosts in this order:

1. Explicit host assignment (`bluetongue.placement.host`)
2. Host group membership
3. Constraint matching
4. Placement strategy
5. Available resources

If a service cannot be placed due to constraints or resource availability, the deployment will fail with an error message explaining why. 
