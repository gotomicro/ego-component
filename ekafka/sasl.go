package ekafka

import (
	"errors"

	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func NewMechanism(saslMechanism, saslUserName, saslPassword string) (sasl.Mechanism, error) {
	var mechanism sasl.Mechanism
	var err error
	if saslMechanism != "" {
		switch saslMechanism {
		case "SCRAM-SHA-256":
			mechanism, err = scram.Mechanism(scram.SHA256, saslUserName, saslPassword)
			if err != nil {
				return nil, err
			}
		case "SCRAM-SHA-512":
			mechanism, err = scram.Mechanism(scram.SHA512, saslUserName, saslPassword)
			if err != nil {
				return nil, err
			}
		case "PLAIN":
			mechanism = plain.Mechanism{
				Username: saslUserName,
				Password: saslPassword,
			}
		default:
			return nil, errors.New("unknown mechanism")
		}
	}
	return mechanism, nil
}
