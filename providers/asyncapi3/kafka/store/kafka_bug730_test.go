package store_test

/*func TestKafka_Bug773(t *testing.T) {
	var c *asyncapi3.Config
	err := yaml.Unmarshal([]byte(cfg), &c)
	require.NoError(t, err)
	err = c.Parse(&dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka-clusters.yaml")}}, &dynamictest.Reader{})
	require.NoError(t, err)

	s := store.New(c, enginetest.NewEngine(), &eventstest.Handler{})
	var brokers []*kafka.Server
	var clients []*kafkatest.Client
	var servers []*asyncapi3.Server
	for _, server := range c.Servers {
		servers = append(servers, server.Value)
	}
	slices.SortFunc(servers, func(a, b *asyncapi3.Server) int {
		return strings.Compare(a.Host, b.Host)
	})
	for i, server := range servers {
		u := try.MustUrl("//" + server.Host)
		b := &kafka.Server{
			Addr:    fmt.Sprintf(":%v", u.Port()),
			Handler: s,
		}
		go func() {
			err := b.ListenAndServe()
			if !errors.Is(err, kafka.ErrServerClosed) {
				log.Error(err)
			}
		}()
		brokers = append(brokers, b)
		clients = append(clients, kafkatest.NewClient(b.Addr, fmt.Sprintf("consumer-%v", i)))
	}
	defer func() {
		for _, client := range clients {
			client.Close()
		}
		for _, b := range brokers {
			b.Close()
		}
	}()
	res, err := clients[0].FindCoordinator(3, &findCoordinator.Request{
		Key:     "cluster1-consumer-group",
		KeyType: 0,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "localhost", res.Host)
	require.Equal(t, int32(9092), res.Port)

	res, err = clients[1].FindCoordinator(3, &findCoordinator.Request{
		Key:     "cluster2-consumer-group",
		KeyType: 0,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "localhost", res.Host)
	require.Equal(t, int32(9094), res.Port)

	res, err = clients[0].FindCoordinator(3, &findCoordinator.Request{
		Key:     "cluster2-consumer-group",
		KeyType: 0,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "localhost", res.Host)
	require.Equal(t, int32(9094), res.Port)

	meta, err := clients[1].Metadata(9, &metaData.Request{
		Topics: []metaData.TopicName{{Name: "cluster2-topic"}},
	})
	require.NoError(t, err)
	require.Equal(t, metaData.ResponseTopic{
		Name: "cluster2-topic", Partitions: []metaData.ResponsePartition{
			{
				LeaderId:        int32(1),
				ReplicaNodes:    []int32{},
				IsrNodes:        []int32{},
				OfflineReplicas: []int32{},
			},
		},
	}, meta.Topics[0])
	join, err := clients[1].JoinGroup(7, &joinGroup.Request{
		GroupId:            "cluster2-consumer-group",
		SessionTimeoutMs:   45000,
		RebalanceTimeoutMs: 300000,
		MemberId:           "",
		GroupInstanceId:    "",
		ProtocolType:       "consumer",
		Protocols: []joinGroup.Protocol{
			{
				Name:     "range",
				MetaData: []byte("000300000001000e636c7573746572322d746f706963ffffffff00000000ffffffffffff"),
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "", join)
}

const cfg = `
asyncapi: 3.0.0
info:
  title: Kafka Multi-Cluster Simulation
  description: AsyncAPI 3.0.0 specification for simulating two Kafka clusters with Mokapi
  version: 1.0.0

servers:
  kafka-cluster1:
    host: localhost:9092
    protocol: kafka
    description: Kafka Cluster 1 for user events
    tags:
      - name: cluster1
        description: Primary cluster for user-related events

  kafka-cluster2:
    host: localhost:9094
    protocol: kafka
    description: Kafka Cluster 2 for order events
    tags:
      - name: cluster2
        description: Secondary cluster for order-related events

channels:
  cluster1-topic:
    address: cluster1-topic
    messages:
      userEvent:
        $ref: '#/components/messages/UserEvent'
    servers:
      - $ref: '#/servers/kafka-cluster1'

  cluster2-topic:
    address: cluster2-topic
    messages:
      orderEvent:
        $ref: '#/components/messages/OrderEvent'
    servers:
      - $ref: '#/servers/kafka-cluster2'

operations:
  publishUserEvent:
    action: send
    channel:
      $ref: '#/channels/cluster1-topic'
    messages:
      - $ref: '#/components/messages/UserEvent'

  publishOrderEvent:
    action: send
    channel:
      $ref: '#/channels/cluster2-topic'
    messages:
      - $ref: '#/components/messages/OrderEvent'

components:
  messages:
    UserEvent:
      name: UserEvent
      title: User Event Message
      summary: Event triggered by user actions
      contentType: application/json
      payload:
        $ref: '#/components/schemas/UserEventPayload'
      examples:
        - name: loginEvent
          summary: User login event
          payload:
            messageId: "msg-001"
            timestamp: "2025-11-06T10:00:00Z"
            userId: "user-123"
            action: "login"
            metadata:
              ip: "192.168.1.1"
              userAgent: "Mozilla/5.0"
      x-mokapi-producer:
        interval: 10s
        count: -1

    OrderEvent:
      name: OrderEvent
      title: Order Event Message
      summary: Event triggered by order operations
      contentType: application/json
      payload:
        $ref: '#/components/schemas/OrderEventPayload'
      examples:
        - name: orderCreated
          summary: Order creation event
          payload:
            eventType: "order.created"
            orderId: "order-456"
            customerId: "customer-789"
            data:
              amount: 99.99
              currency: "USD"
              items:
                - productId: "prod-001"
                  quantity: 2
                  price: 49.99
            timestamp: "2025-11-06T10:05:00Z"
      x-mokapi-producer:
        interval: 15s
        count: -1

  schemas:
    UserEventPayload:
      type: object
      required:
        - messageId
        - timestamp
        - userId
        - action
      properties:
        messageId:
          type: string
          description: Unique identifier for the message
          example: "msg-001"
        timestamp:
          type: string
          format: date-time
          description: ISO 8601 timestamp of the event
          example: "2025-11-06T10:00:00Z"
        userId:
          type: string
          description: Unique identifier for the user
          example: "user-123"
        action:
          type: string
          description: Action performed by the user
          enum: ["login", "logout", "signup", "profile_update"]
          example: "login"
        metadata:
          type: object
          description: Additional event metadata
          properties:
            ip:
              type: string
              description: IP address of the user
              example: "192.168.1.1"
            userAgent:
              type: string
              description: User agent string
              example: "Mozilla/5.0"
          required:
            - ip

    OrderEventPayload:
      type: object
      required:
        - eventType
        - orderId
        - customerId
        - timestamp
      properties:
        eventType:
          type: string
          description: Type of order event
          enum: ["order.created", "order.updated", "order.cancelled", "order.completed"]
          example: "order.created"
        orderId:
          type: string
          description: Unique identifier for the order
          example: "order-456"
        customerId:
          type: string
          description: Unique identifier for the customer
          example: "customer-789"
        data:
          type: object
          description: Order data payload
          properties:
            amount:
              type: number
              description: Total order amount
              minimum: 0
              example: 99.99
            currency:
              type: string
              description: Currency code (ISO 4217)
              pattern: "^[A-Z]{3}$"
              example: "USD"
            items:
              type: array
              description: List of order items
              items:
                type: object
                properties:
                  productId:
                    type: string
                    description: Product identifier
                    example: "prod-001"
                  quantity:
                    type: integer
                    description: Quantity ordered
                    minimum: 1
                    example: 2
                  price:
                    type: number
                    description: Unit price
                    minimum: 0
                    example: 49.99
                required:
                  - productId
                  - quantity
                  - price
          required:
            - amount
            - currency
        timestamp:
          type: string
          format: date-time
          description: ISO 8601 timestamp of the event
          example: "2025-11-06T10:05:00Z"

x-mokapi:
  kafka:
    clusters:
      - name: cluster1
        brokers: ["localhost:9092"]
        topics:
          - name: cluster1-topic
            partitions: 3
            replicationFactor: 1
            config:
              cleanup.policy: delete
              retention.ms: 604800000
      - name: cluster2
        brokers: ["localhost:9094"]
        topics:
          - name: cluster2-topic
            partitions: 2
            replicationFactor: 1
            config:
              cleanup.policy: delete
              retention.ms: 604800000
`
*/
