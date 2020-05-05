package versioning

import "testing"

func TestSameVersion(t *testing.T) {
	status := DetermineLifeCycleStatus("1.0.0", "1.0.0")
	if status != Same {
		t.Errorf("Expected [%s], but got [%s]", Same, status)
	}
}

func TestCurrentVersionHigher(t *testing.T) {
	status := DetermineLifeCycleStatus("1.0.0", "2.0.0")
	if status != Unknown {
		t.Errorf("Expected [%s], but got [%s]", Unknown, status)
	}
}

func TestMajorVersionHigher(t *testing.T) {
	status := DetermineLifeCycleStatus("2.0.0", "1.0.0")
	if status != Major {
		t.Errorf("Expected [%s], but got [%s]", Major, status)
	}
}

func TestMinorVersionHigher(t *testing.T) {
	status := DetermineLifeCycleStatus("1.1.0", "1.0.0")
	if status != Minor {
		t.Errorf("Expected [%s], but got [%s]", Minor, status)
	}
}

func TestPatchVersionHigher(t *testing.T) {
	status := DetermineLifeCycleStatus("1.0.1", "1.0.0")
	if status != Patch {
		t.Errorf("Expected [%s], but got [%s]", Patch, status)
	}
}

func TestLatestVersionExtra(t *testing.T) {
	status := DetermineLifeCycleStatus("1.0.0", "1.0.0-5")
	if status != Unknown {
		t.Errorf("Expected [%s], but got [%s]", Unknown, status)
	}
}
