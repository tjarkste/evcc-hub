package api

import (
	"testing"

	"evcc-cloud/backend/internal/storage"
)

func TestCheckACL_SiteCredentials(t *testing.T) {
	prefix := "user/u1/site/s1/evcc"

	tests := []struct {
		name   string
		topic  string
		acc    int
		expect bool
	}{
		{"read own data", "user/u1/site/s1/evcc/site/pvPower", 1, true},
		{"write own data", "user/u1/site/s1/evcc/loadpoints/1/mode", 2, true},
		{"write own /set", "user/u1/site/s1/evcc/loadpoints/1/mode/set", 2, true},
		{"read+write own data", "user/u1/site/s1/evcc/site/pvPower", 3, true},
		{"deny other site", "user/u1/site/s2/evcc/site/pvPower", 1, false},
		{"deny other user", "user/u2/site/s1/evcc/site/pvPower", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckACL(storage.MQTTCredSite, prefix, "u1", tt.topic, tt.acc)
			if got != tt.expect {
				t.Errorf("CheckACL(%q, %q, %d) = %v, want %v", prefix, tt.topic, tt.acc, got, tt.expect)
			}
		})
	}
}

func TestCheckACL_UserCredentials(t *testing.T) {
	tests := []struct {
		name   string
		topic  string
		acc    int
		expect bool
	}{
		{"read own site", "user/u1/site/s1/evcc/site/pvPower", 1, true},
		{"subscribe own site", "user/u1/site/s1/evcc/#", 4, true},
		{"read another own site", "user/u1/site/s2/evcc/site/pvPower", 1, true},
		{"write /set topic", "user/u1/site/s1/evcc/loadpoints/1/mode/set", 2, true},
		{"deny write non-set topic", "user/u1/site/s1/evcc/loadpoints/1/mode", 2, false},
		{"deny read other user", "user/u2/site/s1/evcc/site/pvPower", 1, false},
		{"deny write other user /set", "user/u2/site/s1/evcc/loadpoints/1/mode/set", 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckACL(storage.MQTTCredUser, "user/u1/site", "u1", tt.topic, tt.acc)
			if got != tt.expect {
				t.Errorf("CheckACL(user, %q, %d) = %v, want %v", tt.topic, tt.acc, got, tt.expect)
			}
		})
	}
}
