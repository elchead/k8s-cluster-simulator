package jobparser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCsv(t *testing.T) {
	file := strings.NewReader(`"Date","o10n-worker-l-jbmbp-qsjc6 worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-f88p8-z9hwl worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-rhcpr-g2pjm worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-45pnm-ghd49 worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-kls5k-tw7jj worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-qpv88-w69cl worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-ghvqt-klrmb worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-879vd-qp7s6 worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-94qhk-vxdxz worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-wnlfd-dv594 worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-7j7bg-gqzgl worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-c5cjd-x8l58 worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-zgqdd-h4fwl worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-5r744-sq5pn worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-s4dtc-l6k47 worker (builtin:containers.memory.residentSetBytes) [B]"
"2022-05-11 08:00:00",21571718826.5,3524325375.5,54965999615.5,,,,,,,16074645503.5,22920224426.5,,,,`)
	podmemories := ParsePodMemories(file)
	assert.Equal(t, "o10n-worker-l-jbmbp-qsjc6", podmemories[0].Name)
	assert.Equal(t, "2022-05-11 08:00:00", podmemories[0].Records[0].Time)
	assert.Equal(t, 21571718826.5, podmemories[0].Records[0].Usage)

	assert.Equal(t, "2022-05-11 08:00:00", podmemories[1].Records[0].Time)
	assert.Equal(t, "o10n-worker-l-f88p8-z9hwl", podmemories[1].Name)
	assert.Equal(t, 3524325375.5, podmemories[1].Records[0].Usage)
}
