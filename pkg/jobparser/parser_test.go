package jobparser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCsv(t *testing.T) {
	file := strings.NewReader(`"Date","o10n-worker-l-jbmbp-qsjc6 worker (builtin:containers.memory.residentSetBytes) [B]","o10n-worker-l-f88p8-z9hwl worker (builtin:containers.memory.residentSetBytes) [B]"`)
	podmemories := ParsePodMemories(file)
	assert.Equal(t, "o10n-worker-l-jbmbp-qsjc6", podmemories[0].Name)
	assert.Equal(t, "o10n-worker-l-f88p8-z9hwl", podmemories[1].Name)
}
