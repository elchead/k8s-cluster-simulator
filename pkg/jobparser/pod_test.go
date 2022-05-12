package jobparser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPodSpecFromPodMemory(t *testing.T) {
	podmem := PodMemory{Name: "w1", Records: []Record{{Time: time.Now(), Usage: 1e9}, {Time: time.Now().Add(2 * time.Minute), Usage: 1e2}}}
	podspec := CreatePod(podmem)
	assert.Equal(t, "\n- seconds: 0.000000\n  resourceUsage:\n    cpu: 8\n    memory: 1000000000.000000\n\n- seconds: 120.000000\n  resourceUsage:\n    cpu: 8\n    memory: 100.000000\n", podspec.Annotations["simSpec"])
}
