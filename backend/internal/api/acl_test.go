package api

import (
	"testing"
)

const testUserID = "550e8400-e29b-41d4-a716-446655440000"
const testPrefix = "user/" + testUserID + "/evcc"

func TestCheckACL_ReadOwnPrefix(t *testing.T) {
	// Reading the exact prefix is allowed.
	if !CheckACL(testPrefix, testPrefix, 1) {
		t.Error("expected read on own prefix to be allowed")
	}
}

func TestCheckACL_ReadOwnSubtopic(t *testing.T) {
	cases := []string{
		testPrefix + "/site/pvPower",
		testPrefix + "/loadpoints/1/mode",
		testPrefix + "/loadpoints/1/chargePower",
	}
	for _, topic := range cases {
		if !CheckACL(testPrefix, topic, 1) {
			t.Errorf("expected read on %q to be allowed", topic)
		}
	}
}

func TestCheckACL_WriteSetTopic(t *testing.T) {
	cases := []string{
		testPrefix + "/loadpoints/1/mode/set",
		testPrefix + "/loadpoints/1/minSoc/set",
		testPrefix + "/site/set",
	}
	for _, topic := range cases {
		if !CheckACL(testPrefix, topic, 2) {
			t.Errorf("expected write on %q to be allowed", topic)
		}
	}
}

func TestCheckACL_WriteNonSetTopicDenied(t *testing.T) {
	cases := []string{
		testPrefix + "/site/pvPower",
		testPrefix + "/loadpoints/1/mode",
	}
	for _, topic := range cases {
		if CheckACL(testPrefix, topic, 2) {
			t.Errorf("expected write on %q to be denied (no /set suffix)", topic)
		}
	}
}

func TestCheckACL_CrossUserDenied(t *testing.T) {
	otherPrefix := "user/other-user-id/evcc"
	if CheckACL(testPrefix, otherPrefix+"/site/pvPower", 1) {
		t.Error("expected cross-user read to be denied")
	}
	if CheckACL(testPrefix, otherPrefix+"/loadpoints/1/mode/set", 2) {
		t.Error("expected cross-user write to be denied")
	}
}

func TestCheckACL_UnknownAccDenied(t *testing.T) {
	if CheckACL(testPrefix, testPrefix+"/site/pvPower", 99) {
		t.Error("expected unknown acc to be denied")
	}
}

func TestCheckACL_ReadWriteCombined(t *testing.T) {
	// acc=3 (read+write): only /set topics should be allowed.
	if !CheckACL(testPrefix, testPrefix+"/loadpoints/1/mode/set", 3) {
		t.Error("expected acc=3 write+read on /set topic to be allowed")
	}
	if CheckACL(testPrefix, testPrefix+"/site/pvPower", 3) {
		t.Error("expected acc=3 on non-set topic to be denied")
	}
}
