package asyncapi3test

import "mokapi/providers/asyncapi3"

type ServerOptions func(s *asyncapi3.Server)

func WithServerDescription(description string) ServerOptions {
	return func(s *asyncapi3.Server) {
		s.Description = description
	}
}

func WithServerTags(tags ...asyncapi3.Tag) ServerOptions {
	return func(s *asyncapi3.Server) {
		for _, tag := range tags {
			s.Tags = append(s.Tags, &asyncapi3.TagRef{Value: &tag})
		}
	}
}

func WithKafkaBinding(key, value string) ServerOptions {
	return func(s *asyncapi3.Server) {
		s.Bindings.Kafka.Config[key] = value
	}
}
