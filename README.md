 
 [![Build Status](https://travis-ci.com/xente/caronte.svg?token=p37nQ84eR3mMdfKzHf6B&branch=master)](https://travis-ci.com/xente/caronte)

# caronte
Caronte has been created to provide swarm services horizontal 
based threshold for metrics.

## Features

- Configuration automatically 
- Supports multiple infrastructure providers (current only AWS)
- Support multiple metrics stores providers (cloudWatch, prometheus)
- Scale rules defined by services 

 ## Configuration
 
 |  command |   Description |
 |---|---|
 | log.level | Set log level. Default value INFO. Allowed DEBUG and INFO |
 | dashboard | Activate Caronte dashboard |
 | dashboard.port | Define Caronte dashboard port. Default value 80 |
 | service.scheduler.discovery.time | Define service discovery timer in seconds |
 | sqs.metic.publisher.queue.name | Activate AWS SQS metrcis |
 | sqs.metic.publisher.queue.time | Define AWS SQS metrics time |

 ## Service Configuration labels

 |  Level |  Scope |  Description |
 |---|---|---|
 | caronte.scale.max   |  Service | Define the max replicas for a target service  |
 | caronte.scale.min   |  Service |  Define the min replicas for a target service |
 | caronte.scale.step  |  Service |  Define the step replicas increase for a target service  |
 | caronte.scale.service.coolDownDelay | Service  | Define coolDown time in seconds for services scale |
 | caronte.scale.maxPreplicasPerNode | Service | Define max replicas per node when the instance provider is activated |
 | caronte.metric.store  | Metrics  |  Metric store to be used allowed (cloudwatch , prometheus)  |
 | caronte.metric.query | Metrics | Metric store query |
 | caronte.metric.scaleUpThreshold  |  Metrics | Scale up metric Threshold   |
 | caronte.metric.scaleDownThreshold |  Metrics |  Scale down metric Threshold |
 | caronte.metric.prometheus.address | Metrics/Prometheus  | Prometheus server address  |
 | caronte.metric.aws.period | Metrics/AWS | CloudWatch query period in seconds  |
 | caronte.instance.provider | Instances | Instances provider allowed (aws) |
 | caronte.instance.coolDownDelay | Instances | Define coolDown delay time in seconds for Instance  |
 | caronte.instance.aws.asg.filters | Instances/aws | Tags filters to define Aws AutoscalingGroup   |
 
 ## Configuration Sample
 
 Scale based on Aws - CloudWatch metric store
 ```yaml
  my-service:
     image: my-service
     deploy:
       replicas: 1
       labels:
         caronte.enable: "true"
         caronte.scale.max: 8
         caronte.scale.min: 1
         caronte.scale.step: 2
         caronte.scale.service.coolDown: 10
         caronte.metric.scaleDownThreshold: 0
         caronte.metric.scaleUpThreshold: 0
         caronte.metric.store: "cloudwatch"
         caronte.metric.query: "SEARCH('{AWS/SQS,QueueName}my-queue-name MetricName=\"NumberOfMessagesDeleted\"', 'Average', 300)"
         caronte.metric.aws.period: 60
         caronte.instance.provider: "aws"
         caronte.instance.aws.asg.filters: '[{"Name":"key", "Values":["my-asg-tag-name"]}]'
   ```
 Scale based on Prometheus metric store
 ```yaml
  my-service-promehteus-provider:
       image: my-service
       deploy:
         replicas: 1
         labels:
           caronte.enable: "true"
           caronte.scale.max: 8
           caronte.scale.min: 1
           caronte.scale.step: 2
           caronte.scale.service.coolDown: 10
           caronte.metric.scaleDownThreshold: 0
           caronte.metric.scaleUpThreshold: 0
           caronte.metric.store: "prometheus"
           caronte.metric.prometheus.address:  "http://localhost:9090"
           caronte.metric.query: "rate(prometheus_tsdb_head_samples_appended_total[5m])"
           caronte.instance.provider: "aws"
           caronte.instance.aws.asg.filters: '[{"Name":"value", "Values":["my-asg-tag-value"]}]'
  ```

## Installation 
Add Caronte as a swarm service.

 ```yaml
  caronte:
    image: xente/caronte
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    restart: always
    deploy:
      replicas: 1
      placement:
        constraints:
          - "node.role==manager"
  ```

If you want to use Caronte with AWS instance provider you have to provide the AWS keys and Region.
```yaml
  caronte:
    image: xente/caronte
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment: 
     - AWS_ACCESS_KEY_ID=
     - AWS_SECRET_ACCESS_KEY=
     - AWS_DEFAULT_REGION=
    restart: always
    deploy:
      replicas: 1
      placement:
        constraints:
          - "node.role==manager"
  ```
