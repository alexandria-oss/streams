# Streams Driver for Amazon Messaging Services

The **stream driver** for `Amazon` messaging services offers both `Writer` and `Reader` implementations through services such as **_Amazon Simple Notification Service (SNS)_** and **_Amazon Simple Queue Service (SQS)_**.

Both `Writer` implementations share a base writer instance which encapsulates a **concurrent batching buffering mechanism** to enable message _batch writing_ capabilities with _high throughput_.

## Amazon Simple Notification Service

For this driver, only `Writer` implementation is available for use. A `Reader` implementation is not on the roadmap as SNS does not have polling mechanisms; instead, the service makes synchronous request directly to subscribers.

This built-in mechanism erases the need for a `Reader` implementation.

## Amazon Simple Queue Service

This driver offers both `Reader` and `Writer` implementations.

## Topic-Queue Chaining Pattern
 
The topic queue chaining pattern is a messaging pattern that can be used to decouple microservices. In this pattern, a topic is used to publish messages to a group of subscribers. Each subscriber is subscribed to the topic, but the messages are delivered to the subscribers individually. This allows the subscribers to process the messages in parallel, which can improve performance.

The topic queue chaining pattern is often used in conjunction with a message queue. In this case, the topic is used to publish messages to the message queue. The message queue then delivers the messages to the subscribers. This allows the subscribers to process the messages in a reliable and scalable way.

The topic queue chaining pattern is a powerful tool that can be used to improve the performance and scalability of microservices. It is a good choice for systems that need to be able to process a large number of messages in a short amount of time.

In other words, the pattern refers to adding a queue, in this case an SQS queue, between the SNS topic and each of the subscriber services/workers. As messages are buffered persistently in an SQS queue, it prevents lost messages if a subscriber process run into problems for many hours or days.

This driver offers the posibility to implement the **topic-queue chaining** pattern. Systems should use the Amazon SNS `Writer` implementation for message publishing and the Amazon SQS `Reader` implementation for consuming messages on their respective queues at subscriber/worker level.

![Topic-Queue chaining pattern diagram](https://d2908q01vomqb2.cloudfront.net/1b6453892473a467d07372d45eb05abc2031647a/2019/11/21/Pic8-1024x418.jpg "Topic-Queue chaining pattern diagram")

### Queues as buffering load balancers

An SQS queue in front of each subscriber service also acts as a buffering load balancer. This pattern is also called `Queue-Based Load Leveling` ([reference doc](https://learn.microsoft.com/en-us/azure/architecture/patterns/queue-based-load-leveling)).

Since every message is delivered to one of potentially many consumer processes, you can scale out the subscriber services, and the message load is distributed over the available consumer processes.

As messages are buffered in the queue, they are preserved during a scaling event, such as when you must wait until an additional consumer process becomes operational.

Moreover, this pattern may also used directly (_no SNS topics required_) if you are only insterested in load balance your application tasks (_e.g. video processing, data analytics/enrichment_).

The `Writer` implementation could be used as underlying communication component for the **asynchronous command bus** if the system implements the **Command and Query Responsibility Segretation (CQRS)** pattern.

Lastly, these queue characteristics help flatten peak loads for your consumer processes, buffering messages until consumers are available. This allows you to process messages at a pace decoupled from the message source.
