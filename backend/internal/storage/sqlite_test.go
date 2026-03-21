package storage

import (
	"os"
	"testing"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	path := t.TempDir() + "/test.db"
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(path)
	})
	return db
}

func TestCreateSite(t *testing.T) {
	db := setupTestDB(t)
	user, err := db.CreateUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	site, err := db.CreateSite(user.ID, "Vacation House", nil)
	if err != nil {
		t.Fatalf("create site: %v", err)
	}
	if site.Name != "Vacation House" {
		t.Errorf("expected name 'Vacation House', got %q", site.Name)
	}
	if site.UserID != user.ID {
		t.Errorf("expected userID %s, got %s", user.ID, site.UserID)
	}
	if site.TopicPrefix == "" {
		t.Error("expected non-empty topic prefix")
	}
	if site.MQTTUsername == "" || site.MQTTPassword == "" {
		t.Error("expected non-empty MQTT credentials")
	}
}

func TestGetSitesByUserID(t *testing.T) {
	db := setupTestDB(t)
	user, _ := db.CreateUser("test@example.com", "password123")

	// CreateUser calls EnsureDefaultSite, so we should have 1 site
	sites, err := db.GetSitesByUserID(user.ID)
	if err != nil {
		t.Fatalf("get sites: %v", err)
	}
	if len(sites) != 1 {
		t.Fatalf("expected 1 default site, got %d", len(sites))
	}
	if sites[0].Name != "My Home" {
		t.Errorf("expected default site name 'My Home', got %q", sites[0].Name)
	}
}

func TestDeleteSite(t *testing.T) {
	db := setupTestDB(t)
	user, _ := db.CreateUser("test@example.com", "password123")
	site, _ := db.CreateSite(user.ID, "To Delete", nil)

	err := db.DeleteSite(site.ID, user.ID)
	if err != nil {
		t.Fatalf("delete site: %v", err)
	}

	sites, _ := db.GetSitesByUserID(user.ID)
	for _, s := range sites {
		if s.ID == site.ID {
			t.Error("site was not deleted")
		}
	}
}

func TestDeleteSiteWrongUser(t *testing.T) {
	db := setupTestDB(t)
	user1, _ := db.CreateUser("user1@example.com", "password123")
	user2, _ := db.CreateUser("user2@example.com", "password123")
	site, _ := db.CreateSite(user1.ID, "User1 Site", nil)

	err := db.DeleteSite(site.ID, user2.ID)
	if err == nil {
		t.Error("expected error when deleting another user's site")
	}
}

func TestMigrationCreatesDefaultSite(t *testing.T) {
	db := setupTestDB(t)
	user, _ := db.CreateUser("test@example.com", "password123")
	sites, _ := db.GetSitesByUserID(user.ID)
	if len(sites) != 1 {
		t.Fatalf("expected 1 default site after creation, got %d", len(sites))
	}
	if sites[0].Name != "My Home" {
		t.Errorf("expected 'My Home', got %q", sites[0].Name)
	}
	expected := "user/" + user.ID + "/site/" + sites[0].ID + "/evcc"
	if sites[0].TopicPrefix != expected {
		t.Errorf("expected topic prefix %q, got %q", expected, sites[0].TopicPrefix)
	}
}

func TestGetSiteByMQTTUsername(t *testing.T) {
	db := setupTestDB(t)
	user, _ := db.CreateUser("test@example.com", "password123")
	site, _ := db.CreateSite(user.ID, "Test Site", nil)

	found, err := db.GetSiteByMQTTUsername(site.MQTTUsername)
	if err != nil {
		t.Fatalf("get site by mqtt username: %v", err)
	}
	if found.ID != site.ID {
		t.Errorf("expected site %s, got %s", site.ID, found.ID)
	}
}
