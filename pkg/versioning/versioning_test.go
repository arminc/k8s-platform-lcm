package versioning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHighestOnlyOne(t *testing.T) {
	version, _ := FindHighestSemVer([]string{"0.0.0"})
	assert.Equal(t, "0.0.0", version, "Does not match only one result")
}

func TestHighestNone(t *testing.T) {
	_, err := FindHighestSemVer([]string{})
	assert.Error(t, err)
}

func TestHighestNotSemVer(t *testing.T) {
	_, err := FindHighestSemVer([]string{"0"})
	assert.Error(t, err)
}

func TestHighestMajor(t *testing.T) {
	version, _ := FindHighestSemVer([]string{"1.0.0", "3.0.0", "2.0.0"})
	assert.Equal(t, "3.0.0", version, "Does not match major")
}

func TestHighestMinor(t *testing.T) {
	version, _ := FindHighestSemVer([]string{"0.1.0", "0.3.0", "0.2.0"})
	assert.Equal(t, "0.3.0", version, "Does not match minor")
}

func TestHighestPatch(t *testing.T) {
	version, _ := FindHighestSemVer([]string{"0.0.1", "0.0.3", "0.0.2"})
	assert.Equal(t, "0.0.3", version, "Does not match patch")
}

func TestHighestMultipleSame(t *testing.T) {
	version, _ := FindHighestSemVer([]string{"0.0.1", "0.0.3", "0.0.1", "0.0.2", "0.0.3"})
	assert.Equal(t, "0.0.3", version, "Does not match one of the same")

	version, _ = FindHighestSemVer([]string{"0.0.1", "0.0.3", "1.2.0", "0.0.1", "0.0.2", "0.0.3"})
	assert.Equal(t, "1.2.0", version, "Does not match one of the same")
}

func TestHighestPre(t *testing.T) {
	version, _ := FindHighestSemVer([]string{"0.0.1-alpha.a", "0.0.1-alpha.c", "0.0.1-alpha.b"})
	assert.Equal(t, "0.0.1-alpha.c", version, "Does not match pre")
}

func TestHighestBuild(t *testing.T) {
	version, _ := FindHighestSemVer([]string{"0.0.1+2", "0.0.1+5", "0.0.1+1"})
	assert.Equal(t, "0.0.1+1", version, "Does not match build")
}

func TestHighestPreAndBuild(t *testing.T) {
	version, _ := FindHighestSemVer([]string{"0.0.1-beta.1+2", "0.0.1-beta.2+2", "0.0.1-beta.2+1"})
	assert.Equal(t, "0.0.1-beta.2+1", version, "Does not match pre and build")
}
