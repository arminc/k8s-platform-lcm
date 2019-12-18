package kubernetes

import "testing"

func TestPodStringToPodStructNoVersion(t *testing.T) {
	pod, _ := ImageStringToContainerStruct("test")
	if pod.FullPath != "test" || pod.URL != "docker.io" || pod.Name != "library/test" || pod.Version != "0" {
		t.Errorf("No version %v", pod)
	}
}

func TestPodStringToPodStructLatestVersion(t *testing.T) {
	pod, _ := ImageStringToContainerStruct("test:latest")
	if pod.FullPath != "test:latest" || pod.URL != "docker.io" || pod.Name != "library/test" || pod.Version != "0" {
		t.Errorf("Latest version %v", pod)
	}
}

func TestPodStringToPodStructOtherVersion(t *testing.T) {
	pod, _ := ImageStringToContainerStruct("test:1.3")
	if pod.FullPath != "test:1.3" || pod.URL != "docker.io" || pod.Name != "library/test" || pod.Version != "1.3" {
		t.Errorf("Other version %v", pod)
	}
}

func TestPodStringToPodStructWithUrl(t *testing.T) {
	pod, _ := ImageStringToContainerStruct("gcr.io/test:1.3")
	if pod.FullPath != "gcr.io/test:1.3" || pod.URL != "gcr.io" || pod.Name != "test" || pod.Version != "1.3" {
		t.Errorf("With url %v", pod)
	}
}

func TestPodStringToPodStructWithImageSubname(t *testing.T) {
	pod, _ := ImageStringToContainerStruct("gcr.io/somebody/test:1.3")
	if pod.FullPath != "gcr.io/somebody/test:1.3" || pod.URL != "gcr.io" || pod.Name != "somebody/test" || pod.Version != "1.3" {
		t.Errorf("With image subname %v", pod)
	}
}

func TestPodStringToPodStructWithPort(t *testing.T) {
	pod, _ := ImageStringToContainerStruct("gcr.io:443/somebody/test:1.3")
	if pod.FullPath != "gcr.io:443/somebody/test:1.3" || pod.URL != "gcr.io:443" || pod.Name != "somebody/test" || pod.Version != "1.3" {
		t.Errorf("With port %v", pod)
	}
}
