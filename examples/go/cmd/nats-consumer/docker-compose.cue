package nats_consumer

import (
	"list"
)

let tagNATSServer = "v2.11.0-preview.1-alpine3.19"
let tagNATSBox = "0.14.3"
let tagBenthos = "4"

_defaultNATSClusterOpts: [
	"--cluster", "nats://0.0.0.0:6222",
	"--cluster-name", "test-cluster",
]

_defaultNATSEntrypoint: list.Concat([["nats-server", "-js"], defaultNATSClusterOpts])

#NATS: {
	image: "synadia/nats-server:\(tagNATSServer)"
	entrypoint: *_defaultNATSEntrypoint | [...string]

	depends_on?: [...string]
	ports?: [...string]
}

services: {
	nats1: #NATS
	nats2: #NATS & {
		entrypoint: _defaultNATSEntrypoint
	}
	nats3: #NATS & {}

	box: {
		image:       "natsio/nats-box:\(tagNATSBox)"
		command:     "sleep infinity"
		working_dir: "/work"
		volumes: ["./backup:/work/backup:ro"]
		environment: NATS_URL: "nats://nats:4222/"
	}

	benthos: {
		image:       "jeffail/benthos:\(tagBenthos)"
		working_dir: "/work"
		environment: NATS_URL: "nats://nats1:4222/"
		volumes: ["./seed.yaml:/benthos.yaml:ro"]
	}
}
